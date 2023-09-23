function sendRegionDeleteRequest(region_name) {
    if (region_name == "") {
        return;
    }

    fetch('regions/' + region_name, {
        method: 'DELETE'
    });

    fetch('/', {
        method: 'GET'
    });
}
