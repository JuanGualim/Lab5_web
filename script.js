function toggleTable() {
    var table = document.querySelector("table");

    if (table.style.display === "none") {
        table.style.display = "table";
    } else {
        table.style.display = "none";
    }
}
function nextEpisode(id) {
    console.log("Increment series with id:", id);
}