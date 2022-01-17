const form = document.querySelector("form");
const input = document.querySelector("#usermsg");
const chatbox = document.querySelector("#chatbox");

console.log(location.hostname);
const socket = new WebSocket(`ws://${location.hostname}:${location.port}/ws`);

form.addEventListener("submit", (ev) => {
  ev.preventDefault();

  const msgElem = document.createElement("div");
  msgElem.innerHTML = input.value;

  //chatbox.append(input.value);

  socket.send(input.value);
  input.value = "";
});

socket.onmessage = (ev) => {
  //console.log(ev.data);

  const msgElem = document.createElement("div");
  msgElem.innerHTML = ev.data;

  chatbox.append(msgElem);
};
