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
          pages = Math.ceil(totalRecords/pageSize)
          for (let i = 1; i <= pages; i++) {
            htmlTmpl += `<li class="page-item ${i === currPageIndex ? 'active' : ''}"><a class="page-link" href="#" onclick="updatePage(${pageSize}, ${i})">${i}</a></li>`;
          }
    
          if (currPageIndex == pages){
            htmlTmpl += `<li class="page-item disabled"><a class="page-link" href="#">Next</a></li>`;
          } else{
            htmlTmpl += `<li class="page-item"><a class="page-link" href="#" onclick="updatePage(${pageSize}, ${currPageIndex+1})">Next</a></li>`;
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