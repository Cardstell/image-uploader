var countLoad = 10;
var downloading = false;

function loadElements() {
	if (downloading || startIndex < 0) return;
	downloading = true;
	var xhttp = new XMLHttpRequest();
	xhttp.onreadystatechange = function() {
		if (this.readyState == 4) {
			downloading = false;
			if (this.status != 200) return;
			var response = JSON.parse(xhttp.responseText);
			if (response.ok !== "true") return;
			for (var i = 0;i<response.result.length;++i) {
				var block = "<div class=\"block\"><a href=\"" + response.result[i].url + 
					"\"><img class=\"imageblock\"src=\"" + response.result[i].previewURL + 
					"\"/></a><p class=\"label2\">" + response.result[i].time + "</p></div>";
				document.body.innerHTML += block;
			}
		}
	};
	xhttp.open("POST", "", true);
	xhttp.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
	var formData = new FormData();
	var endIndex = Math.max(startIndex - countLoad - 1, 0);
	xhttp.send("start=" + startIndex + "&" + "end=" + endIndex);
	startIndex = endIndex - 1;
}

function main() {
	window.addEventListener("scroll", function(event) {
		if ((window.innerHeight + window.scrollY) >= document.body.offsetHeight) {
			loadElements();
		}
		// if (wrapper.scrollTop + wrapper.offsetHeight + 100 > content.offsetHeight) {
		// 	console.log("jopa")
		// 	load();
		// }
	}, false);
	loadElements();
}

main();