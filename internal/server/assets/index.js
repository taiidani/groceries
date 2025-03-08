let list = document.getElementById("list")

if (list) {
    list.addEventListener("change", function (evt) {
        var row = evt.target.closest("tr")
        row.classList.toggle("done", evt.target.checked)
    })
}

// Autofocus anything that wants attention on page load
let autofocuses = document.getElementsByClassName("autofocus")
if (autofocuses.length > 0) {
    autofocuses[0].focus()
}
