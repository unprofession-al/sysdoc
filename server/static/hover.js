var connections = document.querySelectorAll('.connection');

connections.forEach((connection) => {
  connection.addEventListener('mouseover', function(event) {
    // console.log("mouseover");
    saveStyles(event.target);
    if (!isClicked(event.target)) {
      setHovered(event.target);
      return
    }
  });
  connection.addEventListener('mouseout', function(event) {
    // console.log("mouseout");
    saveStyles(event.target);
    if (!isClicked(event.target)) {
      resetStyles(event.target);
      return
    }
  });
  connection.addEventListener('click', function(event) {
    saveStyles(event.target);
    if (isClicked(event.target)) {
      // console.log("toggle clicked off");
      resetStyles(event.target);
    } else {
      // console.log("toggle clicked on");
      setClicked(event.target);
    }
  });
});

function saveStyles(e) {
  if (!e.hasAttribute("data-original-stroke")) {
    e.dataset.originalStroke = e.style.stroke;
  } 
  if (!e.hasAttribute("data-original-stroke-width")) {
    e.dataset.originalStrokeWidth = e.style.strokeWidth;
  } 
} 

function resetStyles(e) {
  e.removeAttribute("data-state-hovered");
  e.removeAttribute("data-state-clicked");
  e.style.stroke = e.dataset.originalStroke;
  e.style.strokeWidth = e.dataset.originalStrokeWidth;
} 

function setHovered(e) {
  e.dataset.stateHovered = true;
  e.style.stroke = 'green';
  e.style.strokeWidth = boldStroke(e);
}

function isHovered(e) {
  return e.hasAttribute("data-state-hovered");
}

function setClicked(e) {
  e.dataset.stateClicked = true;
  e.style.stroke = 'red';
  e.style.strokeWidth = boldStroke(e);
}

function isClicked(e) {
  return e.hasAttribute("data-state-clicked");
}

function boldStroke(e) {
  out = e.style.strokeWidth;
  if (!e.dataset.originalStrokeWidth) {
    out = e.style.strokeWith*3;
  } else {
    out = e.dataset.originalStrokeWidth*3;
  }
  return out
}
