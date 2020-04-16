const sendRequest = (data = {}, method = "get") => {
    return new Promise((resolve, reject) => {
        $.ajax({
            url: "https://s1.navinda.xyz/ceylon_news/v2.0/",
            method: method,
            beforeSend: (request) => {
                request.setRequestHeader("token", "");
            },
            dataType: "json",
            data: data,
            timeout: 10000
        }).done((res) => {
            resolve(res);
        }).fail((jqXHR) => {
            hideOutputToast();
            if (jqXHR.responseJSON.error) {
                showTimedToast(jqXHR.responseJSON.error, 3000);
            } else {
                showTimedToast("Request failed!.", 3000);
            }
            reject("Request failed.");
        })
    });
}