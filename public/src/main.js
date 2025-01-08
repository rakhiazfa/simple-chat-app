class SockioEvent {
  name;
  message;

  constructor(name, message) {
    this.name = name;
    this.message = message;
  }

  toString() {
    return JSON.stringify({
      name: this.name,
      message: this.message,
    });
  }
}

class Sockio extends WebSocket {
  events = {};

  constructor(url, protocols) {
    super(url, protocols);

    this.addEventListener("message", ({ data }) => {
      const event = JSON.parse(data);

      this.trigger(event.name, event.message);
    });
  }

  on(event, callback) {
    if (!this.events[event]) {
      this.events[event] = callback;
    }
  }

  emit(event, message) {
    if (this.readyState) {
      this.send(new SockioEvent(event, message).toString());
    } else {
      setTimeout(() => this.emit(event, message), 1000);
    }
  }

  trigger(event, message) {
    if (this.events[event]) {
      this.events[event](message);
    }
  }
}

document.addEventListener("DOMContentLoaded", () => {
  const chatBox = document.getElementById("chat-box");
  const chatForm = document.getElementById("chat-form");
  const messageInput = document.getElementById("message-input");

  const io = new Sockio("ws://localhost:8080/ws");

  var username = prompt("Please enter your name : ") || "anonymous";

  io.emit("join", username);

  io.on("message", (data) => {
    const chat = JSON.parse(data);
    printMessage(chat.from, chat.message);
  });

  chatForm.addEventListener("submit", (e) => {
    e.preventDefault();
    const message = messageInput.value.trim();

    if (message) {
      printMessage(username, message, "right");
      io.emit("message", message);

      messageInput.value = "";
    }
  });

  function printMessage(from, message, position = "left") {
    const chat = document.createElement("div");

    chat.innerHTML = `
      <div class="mb-4 flex ${
        position == "left" ? "justify-start" : "justify-end"
      }">
        <div>
          <span class="text-xs font-medium ml-1">${from}</span>
          <div class="${
            position == "left"
              ? "bg-gray-200 text-black"
              : "bg-blue-500 text-white"
          } p-3 rounded-lg max-w-xs">
            <p class="text-sm">${message}</p>
          </div>
        </div>
      </div>
    `;

    chatBox.appendChild(chat);
    chatBox.scrollTop = chatBox.scrollHeight;
  }
});
