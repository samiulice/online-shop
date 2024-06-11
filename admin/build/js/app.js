function paginator(currPageIndex, pageSize, totalRecords) {
  let pNav = document.getElementById("paginate_nav");
  let pInfo = document.getElementById("paginator_info");
  if (pNav && pInfo) {
    let startID = (currPageIndex - 1) * pageSize + 1;
    let endID = Math.min(startID + pageSize - 1, totalRecords);
    pInfo.innerHTML = `Showing <strong>${startID}</strong> to <strong>${endID}</strong> of <strong>${totalRecords}</strong> entries`;

    let htmlTmpl = ``;


    if (currPageIndex > 1) {
      htmlTmpl += `<li class="page-item"><a class="page-link" href="#" onclick="updatePage(${pageSize}, ${currPageIndex - 1})">Previous</a></li>`;
    } else {
      htmlTmpl += `<li class="page-item disabled"><a class="page-link" href="#">Previous</a></li>`;
    }
    pages = Math.ceil(totalRecords / pageSize)
    for (let i = 1; i <= pages; i++) {
      htmlTmpl += `<li class="page-item ${i === currPageIndex ? 'active' : ''}"><a class="page-link" href="#" onclick="updatePage(${pageSize}, ${i})">${i}</a></li>`;
    }

    if (currPageIndex == pages) {
      htmlTmpl += `<li class="page-item disabled"><a class="page-link" href="#">Next</a></li>`;
    } else {
      htmlTmpl += `<li class="page-item"><a class="page-link" href="#" onclick="updatePage(${pageSize}, ${currPageIndex + 1})">Next</a></li>`;
    }

    pNav.innerHTML = htmlTmpl;
  }
}


function formatDate(time) {
  const input = "2024-06-25T00:00:00Z";

  // Parse the input string into a Date object
  const date = new Date(input);

  // Define arrays for month names and zero-padding for formatting
  const months = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];
  const pad = (num) => num.toString().padStart(2, '0');

  // Extract date components
  const day = pad(date.getDate());
  const month = months[date.getMonth()];
  const year = date.getFullYear();
  const hours = pad(date.getHours());
  const minutes = pad(date.getMinutes());
  const seconds = pad(date.getSeconds());

  // Format the date
  const formattedDate = `${day}-${month}-${year} ${hours}:${minutes}:${seconds}`;

  return formattedDate

}
//print currency into standard format
function formatCurrency(amount, currency) {
  // Convert cents to taka
  let t = parseFloat(amount / 100);
  let c = currency.toUpperCase()

  // Format the currency using toLocaleString method
  return t.toLocaleString("en-" + c.substring(0, 2), {
    style: "currency",
    currency: c,
  });
}

//getMonthName return the month name for month number
function getMonthName(monthNumber) {
  m = parseInt(monthNumber)
  // Array of month names
  var months = [
    "January", "February", "March", "April", "May", "June",
    "July", "August", "September", "October", "November", "December"
  ];

  // Check if the month number is valid
  if (m >= 1 && m <= 12) {
    // Return the corresponding month name
    return months[m - 1];
  } else {
    // Return an error message if the month number is invalid
    return "Invalid month number";
  }
}

function checkError(data, callback) {
  if (data.message == "Invalid authentication credentials") {
    console.log("not logged in");
    signout();
    return;
  }
  callback();
  return;
}

function showErrorInTable() {
  let newRow = tbody.insertRow();
  let newCell = newRow.insertCell();

  document.getElementById("row-actions").classList.add("d-none");
  newRow.classList.add("text-center");

  newCell = newRow.insertCell();
  newCell.setAttribute('colspan', 7);
  newCell.innerHTML = '<b style="color:rgb(255, 0, 76); width: 30%">Internal Server Error</b>';
  return;
}

function signout() {
  localStorage.removeItem("token")
  localStorage.removeItem("token_expiry")
  location.href = "/signout"
}

function checkAuth() {
  if (localStorage.getItem("token") === null) {
    location.href = "/signin"
    return
  } else {
    let token = localStorage.getItem("token");
    console.log(token)
    const myHeader = new Headers();
    myHeader.append("Content-Type", "application/json");
    myHeader.append("Authorization", "Bearer " + token);

    const requestOptions = {
      method: "POST",
      headers: myHeader,
    }

    fetch("{{.API}}/api/is-authenticated", requestOptions)
      .then(response => response.json())
      .then(function (data) {
        if (data.error === true) {
          console.log("not logged in");
          // location.href = "/signin"
        } else {
          console.log("logged in");
        }
      })
  }
}
function goBack() {
  window.location.href = document.referrer;
}

function toggleFullScreen() {
  let toggleBtn = document.getElementById("screen-icon")
  if (!document.fullscreenElement) {
    toggleBtn.classList.remove("glyphicon-fullscreen");
    toggleBtn.classList.add("glyphicon-resize-small");
    document.documentElement.requestFullscreen().catch(err => {
      console.error(`Error attempting to enable full-screen mode: ${err.message} (${err.name})`);
    });
  } else {
    if (document.exitFullscreen) {
      toggleBtn.classList.remove("glyphicon-resize-small");
      toggleBtn.classList.add("glyphicon-fullscreen");
      document.exitFullscreen();
    }
  }
}

function copyToClipboard(cpyBtn, copyText) {
  const textToCopy = getElementContent(copyText);
  const tooltip = document.getElementById(cpyBtn);

  navigator.clipboard.writeText(textToCopy).then(() => {
    console.log('Text copied to clipboard');

    // Show the tooltip
    tooltip.textContent = 'Copied!';

    // Hide the tooltip after 2 seconds
    setTimeout(() => {
      tooltip.innerHTML = '<i class="fa fa-copy"></i>';
    }, 1500);
  }).catch(err => {
    console.error('Error in copying text: ', err);
  });
}

function getElementContent(elementID) {
  const element = document.getElementById(elementID);
  // Check if the element has a 'value' attribute (form elements)
  if (element instanceof HTMLInputElement ||
    element instanceof HTMLTextAreaElement ||
    element instanceof HTMLSelectElement ||
    element instanceof HTMLButtonElement ||
    element instanceof HTMLOptionElement ||
    element instanceof HTMLDataElement ||
    element instanceof HTMLMeterElement ||
    element instanceof HTMLProgressElement) {
    return element.value;
  }
  // For other elements, use 'textContent'
  return element.textContent;
}
