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

  function updateBlock(object) {
    console.log(object);
    const container = document.getElementById("last-blocks-container");

    const oldElement = document.getElementById(`b-${object.newBlock.id}`);
    if (document.body.contains(oldElement)) {
      // Block already known and displayed
      return;
    }

    const div = htmlToElement(`
    <div id="b-${object.newBlock.id}" class="column is-2">
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
    container.prepend(div);

    for (let t of object.newBlock.transactions) {
      updateTransaction(t, false);
    }
  }

  function updateTransaction(object, isPending) {
    console.log(object);
    const container = document.getElementById("transaction-container");
    const oldElement = document.getElementById(`t-${object.id}`);
    if (document.body.contains(oldElement)) {
      container.removeChild(oldElement);
    }

    let statusHTML;
    if (isPending) {
      statusHTML = '<strong class="has-text-warning">Pending</strong>';
    } else {
      statusHTML = '<strong class="has-text-success">In Chain</strong>';
    }
    const hasData = object.data !== undefined;
    const newElement = htmlToElement(`
    <tr id="t-${object.id}">
      <td>${statusHTML}</td>
      <td><a href="/dashboard/transaction?id=${object.id}">${object.id.substring(0, 6)}</a></td>
      <td><a href="/dashboard/account?id=${object.sender}">${object.sender.substring(0, 6)}</a></td>
      <td><a href="/dashboard/account?id=${object.receiver}">${object.receiver.substring(0, 6)}</a></td>
      <td>${object.balance}</td>
      <td>${new Date(object.timeUnixNano / 1000)}</td>
      <td>${hasData ? "Yes" : "No"}</td>
      <td>${object.fee}</td>
    </tr>
    `);
    container.prepend(newElement);
  }

  function parseMessage(message) {
    const object = JSON.parse(message)
    if (object.newBlock !== undefined) {
      updateBlock(object);
    }
    if (object.newTransaction !== undefined) {
      updateTransaction(object.newTransaction, true);
    }
  }

  if (window["WebSocket"]) {
    const protocol = "https:" == document.location.protocol ? "wss://" : "ws://";
    conn = new WebSocket(protocol + document.location.host + "/dashboard/ws");
    conn.onclose = function(evt) {
      alert("Connection to the server was closed.")
    };
    conn.onmessage = function(evt) {
      let messages = evt.data.split('\n');
      for (let i = 0; i < messages.length; i++) {
        parseMessage(messages[i]);
      }
    };
  } else {
    document.body.innerHTML = "<b>Your browser does not support WebSockets.</b>";
  }
};
