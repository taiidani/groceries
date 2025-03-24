// Autofocus anything that wants attention on page load
let autofocuses = document.getElementsByClassName("autofocus")
if (autofocuses.length > 0) {
    autofocuses[0].focus()
}

// HTMX error handling
document.addEventListener("htmx:responseError", function (evt) {
    alert(evt.detail.xhr.responseText)
})
