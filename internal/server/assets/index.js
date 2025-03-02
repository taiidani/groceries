let list = document.getElementById("list")
let rows = list.querySelectorAll("tr")

rows.forEach(elem => {
    elem.addEventListener("change", function (evt) {
        this.classList.toggle("done", evt.target.checked)
        console.log("changed")
    })
})


function allowFinishShopping() {
    var anyDone = false
    rows.forEach(elem => {
        if (elem.classList.contains("done")) {
            anyDone = true
        }
    })

}
