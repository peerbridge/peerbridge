window.onload = function() {
  var conn;

  /**
   * @param {String} HTML representing a single element
   * @return {Element}
   */
  function htmlToElement(html) {
    var template = document.createElement('template');
    html = html.trim(); // Never return a text node of whitespace as the result
    template.innerHTML = html;
    return template.content.firstChild;
  }

  function parseMessage(message) {
    const object = JSON.parse(message)
    if (object.newBlock !== undefined) {
      const container = document.getElementById("last-blocks-container");
      if (container.childNodes.length >= 12) {
        const nodesToRemove = container.childNodes.length - 11;
        for(let i = 0; i < nodesToRemove; i++) {
          container.removeChild(container.childNodes[i]);
        }
      }
      const div = htmlToElement(`
      <div class="column is-2" data-block-id="${object.newBlock.id}">
        <a class="box" href="/dashboard/block?id=${object.newBlock.id}">
          <h5 class="subtitle">
            <strong>${object.newBlock.id.substring(0, 6)}</strong> <span class="has-text-grey">#${object.newBlock.height}</span>
          </h5>
          <div style="background: #${object.newBlock.id.substring(0, 6)}; height: 0.25rem; border-radius: 0.125rem;"></div>
          <p class="pt-4">Published by ${object.newBlock.creator.substring(0, 6)}</p>
          <p>Contains ${object.newBlock.transactions.length} transactions</p>
        </a>
      </div>
      `);
      container.appendChild(div);
    }
  }

  if (window["WebSocket"]) {
    const protocol = "https:" == document.location.protocol ? "wss://" : "ws://";
    conn = new WebSocket(protocol + document.location.host + "/dashboard/ws");
    conn.onclose = function(evt) {
      alert("Connection closed.")
    };
    conn.onmessage = function(evt) {
      var messages = evt.data.split('\n');
      for (var i = 0; i < messages.length; i++) {
        parseMessage(messages[i]);
      }
    };
  } else {
    document.body.innerHTML = "<b>Your browser does not support WebSockets.</b>";
  }
};
