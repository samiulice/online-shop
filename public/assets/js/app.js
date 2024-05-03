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
function togglePasswordVisibility(element_id) {
  var passwordInput = document.getElementById(element_id);
  var passwordIcon = document.getElementById(element_id + "-icon");

  console.log(element_id)
  if (passwordInput.type === "password") {
      passwordInput.type = "text";
      passwordIcon.src = "/public/assets/icons/eye-solid.svg";
  } else {
      passwordInput.type = "password";
      passwordIcon.src = "/public/assets/icons/eye-slash-solid.svg";
  }
}
//End: togglePasswordVisibility