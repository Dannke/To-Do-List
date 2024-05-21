document.querySelectorAll('.edit-btn').forEach(button => {
    button.addEventListener('click', (e) => {
        const listItem = e.target.closest('li');
        const editForm = listItem.querySelector('.edit-container');
        editForm.classList.toggle('active');
    });
});
