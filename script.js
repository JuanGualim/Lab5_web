function toggleTable() {
    var table = document.querySelector("table");

    if (table.style.display === "none") {
        table.style.display = "table";
    } else {
        table.style.display = "none";
    }
}