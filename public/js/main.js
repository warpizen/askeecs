
function getText(elem) {
  /*
  if(elem.tagName === "TEXTAREA" ||
    (elem.tagName === "INPUT" && elem.type === "text")) {
    return elem.value.substring(elem.selectionStart,
                                elem.selectionEnd);
  }
  */
  if(elem.tagName === "TEXTAREA") {
    return elem.value.substring(elem.selectionStart, elem.selectionEnd);
  }
  return null;
}

//$("#toolbarBold").click(function() {
//function toolbarBold() {
  //console.log("toolbarBold clicked")
  /*
  var txt = getText(document.activeElement); 
  var elem = document.activeElement;
  var value = elem.value;
  elem.value = value.slice(0, elem.selectionStart) + "TEST" + 
    value.slice(sel.selectionEnd);
  */
//}
//});

/*
setInterval(function() {
  var txt = getText(document.activeElement);
  console.log(txt === null ? 'no input selected' : txt)
}, 1000);
*/

/*
setInterval(function() {
  if (window.jQuery) { 
    console.log("jquery working");
  } else {
    console.log("not working");
  }
}, 1000);
*/


