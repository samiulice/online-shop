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