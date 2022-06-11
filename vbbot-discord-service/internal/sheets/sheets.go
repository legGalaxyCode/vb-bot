package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"vb-bot/internal/config"
	"vb-bot/pkg/logging"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(cfg *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok := &oauth2.Token{AccessToken: os.Getenv("ACCESS_TOKEN")}
	if !tok.Valid() {
		tok = getTokenFromWeb(cfg)
		saveTokenInFile(config.GetConfig().Filepath.TokenPath, tok)
	}
	return cfg.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	fmt.Println(tok.AccessToken)
	return tok, err
}

// Saves a token to a file path.
func saveTokenInFile(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

type SpreadsheetPushRequest struct {
	SpreadsheetId string        `json:"spreadsheet_id"`
	Range         string        `json:"range"`
	Values        []interface{} `json:"values"`
}

func CreateSheetsAndDrive(filepath string) (*http.Client, *http.Client) {
	// TODO REFACTOR THIS TO ENV
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	sheetsConfig, err := google.JWTConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	//client := getClient(config)
	sheetsClient := sheetsConfig.Client(oauth2.NoContext)

	driveConfig, err := google.JWTConfigFromJSON(b, "https://www.googleapis.com/auth/drive")
	if err != nil {
		log.Fatal(err)
	}

	driveClient := driveConfig.Client(oauth2.NoContext)

	return sheetsClient, driveClient
}

type SpreadsheetData struct {
	SpreadsheetId  string         `json:"SpreadsheetId,omitempty"`
	SpreadsheetUrl string         `json:"SpreadsheetUrl,omitempty"`
	FirstSheetId   int64          `json:"SheetId,omitempty"`
	LastEntryId    int64          `json:"LastEntryId,omitempty"`
	Logger         logging.Logger `json:"-"`
}

type LogMessage struct {
	Number     int64
	UserID     string
	Time       string
	ContentMsg string
	ImageURL   string
}

type SheetCallWrapper struct {
	SheetService              *sheets.Service
	ValuesUpdateRequests      []*sheets.BatchUpdateValuesRequest
	SpreadsheetUpdateRequests []*sheets.Request
	Data                      SpreadsheetData
}

func NewSpreadsheetDataFromJSON(filename string, logger logging.Logger) (SpreadsheetData, error) {
	spreadsheetData := SpreadsheetData{}
	spreadsheetDataJson, err := os.OpenFile(filename, os.O_RDWR, os.ModeAppend)
	defer spreadsheetDataJson.Close()
	if err != nil {
		logger.Println(err)
		return spreadsheetData, err
	}
	err = json.NewDecoder(spreadsheetDataJson).Decode(&spreadsheetData)
	if err != nil {
		logger.Println(err)
		return spreadsheetData, err
	}
	spreadsheetData.Logger = logger

	return spreadsheetData, err
}

func UpdateDataJson(filename string, data SpreadsheetData) error {
	spreadsheetDataJson, err := os.OpenFile(filename, os.O_RDWR, os.ModeAppend)
	if err != nil {
		data.Logger.Println("Error open json file")
		return err
	}
	defer spreadsheetDataJson.Close()
	//err = json.NewDecoder(spreadsheetDataJson).Decode(&data)
	if err != nil {
		data.Logger.Println("Error decode data struct")
		return err
	}
	mar, err := json.Marshal(data)
	if err != nil {
		data.Logger.Println("Error marshal json file")
		return err
	}
	err = spreadsheetDataJson.Truncate(0)
	if err != nil {
		return err
	}
	_, err = spreadsheetDataJson.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = spreadsheetDataJson.Write(mar)
	if err != nil {
		return err
	}
	return err
}

func NewSheet(data SpreadsheetData) (SheetCallWrapper, error) {
	sheetsClient, _ := CreateSheetsAndDrive(config.GetConfig().Filepath.ClientCreditsPath)
	ctx := context.Background()
	service, err := sheets.NewService(ctx, option.WithHTTPClient(sheetsClient))
	if err != nil {
		data.Logger.Warnf(fmt.Sprintf("Error while create SheetCallWrapper entry: %s", err))
	}

	return SheetCallWrapper{
		SheetService:              service,
		ValuesUpdateRequests:      make([]*sheets.BatchUpdateValuesRequest, 0, 10),
		SpreadsheetUpdateRequests: make([]*sheets.Request, 0, 10),
		Data:                      data,
	}, err
}

func (data SpreadsheetData) Close() error {
	err := UpdateDataJson(config.GetConfig().Filepath.SpreadsheetDataPath, data)
	if err != nil {
		return err
	}
	return err
}

func (sheet *SheetCallWrapper) AddValuesUpdateRequest(valueInputOption string, dataRange string, dataValues [][]interface{}, dataMajorDimension string) {
	sheet.ValuesUpdateRequests = append(sheet.ValuesUpdateRequests, &sheets.BatchUpdateValuesRequest{
		ValueInputOption: valueInputOption,
		Data: []*sheets.ValueRange{
			{
				Range:          dataRange,
				Values:         dataValues,
				MajorDimension: dataMajorDimension,
			},
		},
	})
}

func (sheet *SheetCallWrapper) addDimensionPixelSize(dimension string, startIndex int64, endIndex int64, pixelSize int64) {
	sheet.SpreadsheetUpdateRequests = append(sheet.SpreadsheetUpdateRequests, &sheets.Request{
		UpdateDimensionProperties: &sheets.UpdateDimensionPropertiesRequest{
			Range: &sheets.DimensionRange{
				Dimension:  dimension,
				SheetId:    sheet.Data.FirstSheetId,
				StartIndex: startIndex,
				EndIndex:   endIndex,
			},
			Properties: &sheets.DimensionProperties{
				PixelSize: pixelSize,
			},
			Fields: "PixelSize",
		},
	})
}

func (sheet *SheetCallWrapper) SetColumnsWidth(startCol int64, endCol int64, width int64) {
	sheet.addDimensionPixelSize("COLUMNS", startCol, endCol+1, width)
}

func (sheet *SheetCallWrapper) SetColumnWidth(col int64, width int64) {
	sheet.SetColumnsWidth(col, col, width)
}

func (sheet *SheetCallWrapper) SetRowsHeight(startRow int64, endRow int64, height int64) {
	sheet.addDimensionPixelSize("ROWS", startRow, endRow+1, height)
}

func (sheet *SheetCallWrapper) SetRowHeight(row int64, height int64) {
	sheet.SetRowsHeight(row, row, height)
}

func (sheet *SheetCallWrapper) RunRequests() {
	ctx := context.Background()
	var err error
	_, err = sheet.SheetService.Spreadsheets.BatchUpdate(sheet.Data.SpreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: sheet.SpreadsheetUpdateRequests,
	}).Context(ctx).Do()
	if err != nil {
		fmt.Println("butch update values err: ", err.Error())
	}
	for _, req := range sheet.ValuesUpdateRequests {
		_, err = sheet.SheetService.Spreadsheets.Values.BatchUpdate(sheet.Data.SpreadsheetId, req).Context(ctx).Do()
		if err != nil {
			fmt.Println("butch update err: ", err.Error())
		}
	}
	sheet.SpreadsheetUpdateRequests = nil
	sheet.ValuesUpdateRequests = nil
}
