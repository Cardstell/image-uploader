var uploading = false;
var first = true;
function main() {
	var label = document.getElementsByClassName("label")[0];
	var image = document.getElementsByClassName("image")[0]
	var container = document.getElementsByClassName("container")[0]

	window.addEventListener('paste', function(event){
		if (uploading) return;
		if (!first) {
			
		}
		var items = (event.clipboardData || event.originalEvent.clipboardData).items;
		for (index in items) {
			var item = items[index];
			if (item.kind === 'file') {
				var blob = item.getAsFile();
				var reader = new FileReader();
				reader.onload = function(event){
					var template = "data:image/png;base64,";
					var data = event.target.result;
					if (!data.startsWith(template)) return;

					uploading = true;
					first = false;
					container.style.background = "none";
					
					image.style.display = "none";
					label.style.paddingTop = "45px";
					label.innerHTML = "Uploading...";
					var base64data = data.split(",")[1];
					var xhttp = new XMLHttpRequest();
					xhttp.onreadystatechange = function() {
						if (this.readyState == 4) {
							uploading = false;
							if (this.status != 200) {
								label.innerHTML = "Uploading failed!";
								return;
							}
							var response = JSON.parse(xhttp.responseText);
							if (response.result === undefined) {
								label.innerHTML = "Uploading failed!";
								return;
							}
							label.innerHTML = response.result;
							if (!response.ok) return;
							if (response.url === undefined) return;
							image.style.display = "block";
							image.src = response.url;
							label.style.paddingTop = "5px";
						}
					};
					xhttp.open("POST", ".", true);
					var bstr = atob(base64data), byteArray = new Uint8Array(bstr.length);
					for (var i = 0;i<bstr.length;++i)
						byteArray[i] = bstr.charCodeAt(i);
					var formData = new FormData();
					formData.append("file", new Blob([byteArray], {type:"application/octet-stream"}));
					xhttp.send(formData);
				};
				reader.readAsDataURL(blob);
			}
		}
	});
}

main()