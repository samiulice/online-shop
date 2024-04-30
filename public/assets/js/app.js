//Start Homepage Clock
var dialLines = document.getElementsByClassName('diallines');
var clockEl = document.getElementsByClassName('clock')[0];

for (var i = 1; i < 60; i++) {
    clockEl.innerHTML += "<div class='diallines'></div>";
    dialLines[i].style.transform = "rotate(" + 6 * i + "deg)";
}

function clock() {
    var weekday = [
            "Sunday",
            "Monday",
            "Tuesday",
            "Wednesday",
            "Thursday",
            "Friday",
            "Saturday"
        ],
        d = new Date(),
        h = d.getHours(),
        m = d.getMinutes(),
        s = d.getSeconds(),
        date = d.getDate(),
        month = d.getMonth() + 1,
        year = d.getFullYear(),
            
        hDeg = h * 30 + m * (360/720),
        mDeg = m * 6 + s * (360/3600),
        sDeg = s * 6,
        
        hEl = document.querySelector('.hour-hand'),
        mEl = document.querySelector('.minute-hand'),
        sEl = document.querySelector('.second-hand'),
        dateEl = document.querySelector('.date'),
        dayEl = document.querySelector('.day');

        var day = weekday[d.getDay()];

    if(month < 9) {
        month = "0" + month;
    }

    hEl.style.transform = "rotate("+hDeg+"deg)";
    mEl.style.transform = "rotate("+mDeg+"deg)";
    sEl.style.transform = "rotate("+sDeg+"deg)";
    dateEl.innerHTML = date+"/"+month+"/"+year;
    dayEl.innerHTML = day;
}

setInterval("clock()", 100);
//End Homepage Clock

//start: Toggle nav functions
function toggleNav() {
    var navLinks = document.querySelector('.nav-links');
    var overlay = document.getElementById('overlay');

    navLinks.classList.toggle('active');
    overlay.style.display = (overlay.style.display === "block") ? "none" : "block";
  }

  function closeNav() {
    var navLinks = document.querySelector('.nav-links');
    var overlay = document.getElementById('overlay');

    navLinks.classList.remove('active');
    overlay.style.display = "none";
  }

  window.addEventListener('resize', function() {
    var navLinks = document.querySelector('.nav-links');
    var overlay = document.getElementById('overlay');

    if (window.innerWidth > 768) {
      navLinks.classList.remove('active');
      overlay.style.display = "none";
    }
  });
//End: Toggle nav functions

//Start: togglePasswordVisibility toggle password between show and hide
function togglePasswordVisibility() {
  var passwordInput = document.getElementById("password");
  var passwordIcon = document.getElementById("password-icon");

  if (passwordInput.type === "password") {
      passwordInput.type = "text";
      passwordIcon.src = "/public/assets/icons/eye-solid.svg";
  } else {
      passwordInput.type = "password";
      passwordIcon.src = "/public/assets/icons/eye-slash-solid.svg";
  }
}
//End: togglePasswordVisibility