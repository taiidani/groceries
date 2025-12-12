let itemAdderForm = document.getElementById("itemAdderForm");

// HTMX error handling
document.addEventListener("htmx:responseError", function (evt) {
  alert(evt.detail.xhr.responseText);
});

if (itemAdderForm) {
  itemAdderForm.addEventListener("htmx:responseError", function (evt) {
    evt.stopPropagation();
    evt.target.name.setCustomValidity(evt.detail.xhr.responseText);
    evt.target.reportValidity();
  });
}

// Item adder form
if (itemAdderForm) {
  itemAdderForm.name.addEventListener("keydown", function (evt) {
    // Reset the validity so that modifications reset prior invalid server responses
    evt.target.setCustomValidity("");
  });
}

// Search forms
document.querySelectorAll("[role=search]").forEach(function (form) {
  form.addEventListener("submit", function (evt) {
    evt.preventDefault();
  });

  form.addEventListener("keyup", function (evt) {
    let target = form.getAttribute("search-target");
    let search = form.elements["search"].value.toLowerCase();

    console.log(target);
    console.log(search);

    if (search === "") {
      document.querySelectorAll(target + " .item").forEach(function (n) {
        n.classList.remove("hide");
      });
      return;
    }

    document.querySelectorAll(target + " .item .name").forEach(function (n) {
      let content = n.textContent.toLowerCase();

      if (content.indexOf(search) === -1) {
        n.closest(".item").classList.add("hide");
      } else {
        n.closest(".item").classList.remove("hide");
      }
    });
  });
});
