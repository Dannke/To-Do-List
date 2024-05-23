document.addEventListener("DOMContentLoaded", function() {
    const editButtons = document.querySelectorAll(".edit-btn");
    const editForms = document.querySelectorAll(".edit-container");

    editButtons.forEach((button, index) => {
        button.addEventListener("click", () => {
            editForms[index].style.display = editForms[index].style.display === "none" || editForms[index].style.display === "" ? "flex" : "none";
        });
    });
});
