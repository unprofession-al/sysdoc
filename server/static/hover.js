var connections = document.querySelectorAll('.connection');

connections.forEach((connection) => {
  connection.addEventListener('mouseover', function(event) {
    event.target.style.stroke = 'red';
  });
  connection.addEventListener('mouseout', function(event) {
    event.target.style.stroke = 'black';
  });
});
