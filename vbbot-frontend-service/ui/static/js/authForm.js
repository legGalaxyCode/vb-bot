let loginForm = document.getElementById("loginForm")
let loginDiv = document.getElementById("formDiv")

loginForm.addEventListener("submit", (event) => {
	event.preventDefault();
	let formData = new FormData(event.target);

	const payload = {
		action: formData.get('action'),
		auth: {
			email: formData.get('email'),
			password: formData.get('password'),
			remember: formData.get('remember'),
		}
	}

	const headers = new Headers();
	headers.append("Content-Type", "application/json");

	const request = {
		method: 'POST',
		body: JSON.stringify(payload),
		headers: headers,
	};
	//console.log(payload)

	fetch("http:\/\/localhost:8001/handle", request)
	.then((response) => response.json())
	.then((data) => {
		console.log(data);
		if (data["error"] == false) {
			document.body.innerHTML += '<div class="container"><br><strong style="color: #04AA6D">Authentication success</strong></div>'
		} else {
			document.body.innerHTML += '<div class="container"><br><strong style="color: #f44336">Authentication error</strong></div>'
		}
	})
	.catch((error) => {
		console.log(error);
	})
})
