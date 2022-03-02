const sendRequest = (path, data = {}, method = "get") => {
	return new Promise((resolve, reject) => {
		$.ajax({
			url: "https://arm.navinda.xyz/cn/api/v1.0" + path,
			method: method,
			dataType: "json",
			data: data,
			timeout: 10000,
		})
			.done((res) => {
				resolve(res);
			})
			.fail((jqXHR) => {
				hideOutputToast();
				if (jqXHR.responseJSON.error) {
					showTimedToast(jqXHR.responseJSON.error, 3000);
				} else {
					showTimedToast("Request failed!.", 3000);
				}
				reject("Request failed.");
			});
	});
};
