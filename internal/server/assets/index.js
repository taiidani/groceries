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

// Search forms
document.querySelectorAll("[role=search]").forEach(function (form) {
    form.addEventListener("submit", function (evt) {
        evt.preventDefault()
    })

    form.addEventListener("keyup", function (evt) {
        let target = form.getAttribute("search-target")
        let search = form.elements["search"].value.toLowerCase()

        console.log(target)
        console.log(search)

        if (search === "") {
            document.querySelectorAll(target + " .item").forEach(function (n) {
                n.classList.remove("hide")
            })
            return
        }

        document.querySelectorAll(target + " .item .name").forEach(function (n) {
            let content = n.textContent.toLowerCase()

            if (content.indexOf(search) === -1) {
                n.closest(".item").classList.add("hide")
            } else {
                n.closest(".item").classList.remove("hide")
            }
        })
    })
})
