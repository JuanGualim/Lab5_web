function toggleTable() {
    var table = document.querySelector("table");

    if (table.style.display === "none") {
        table.style.display = "table";
    } else {
        table.style.display = "none";
    }
}
async function nextEpisode(id) {
    const url = `/update?id=${id}`

    const response = await fetch(url, {
        method: "POST"
    })

    location.reload()
}

async function prevEpisode(id) {
    const url = `/decrement?id=${id}`

    const response = await fetch(url, {
        method: "POST"
    })

    location.reload()
}