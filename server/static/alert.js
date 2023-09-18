function displayAlert(msg) {
  const node = document.createElement('div'); 
  node.classList.add('alert');
  const textnode = document.createTextNode(msg);
  node.appendChild(textnode);
  document.getElementById('alert-box').appendChild(node);
}
