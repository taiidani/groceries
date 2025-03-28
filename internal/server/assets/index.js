// HTMX error handling
document.addEventListener("htmx:responseError", function (evt) {
    alert(evt.detail.xhr.responseText)
})

// Modal dialogs
document.querySelectorAll("dialog").forEach(function (dialog) {
    dialog.addEventListener("htmx:afterSwap", function (evt) {
        dialog.showModal()
    })

    dialog.addEventListener("click", function (evt) {
        if (evt.target.getAttribute("role") == "cancel") {
            dialog.close()
        }
    })

    dialog.addEventListener("submit", function (evt) {
        dialog.close()
    })
})
