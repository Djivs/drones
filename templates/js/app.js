function sendRegionDeleteRequest(region_name) {
    if (region_name == "") {
        return;
    }

    // Instantiating new EasyHTTP class
    const http = new DeleteHTTP;

    // Update Post
    http.delete('regions/' + region_name)

    // Resolving promise for response data
    .then(data => console.log(data))

    // Resolving promise for error
    .catch(err => console.log(err))

    .finally(fetch("/", {method: "GET"}));
}
