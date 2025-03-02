let list = document.getElementById("list")

list.addEventListener("change", function (evt) {
    var row = evt.target.closest("tr")
    row.classList.toggle("done", evt.target.checked)
})
