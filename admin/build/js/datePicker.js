//vanila js date picker
document.addEventListener('DOMContentLoaded', function () {
    // Initialize the datepicker
    const elem = document.getElementById('datepickerInput');
    const datepicker = new Datepicker(elem, {
      format: "dd-mm-yyyy",
      showOnFocus: true,
      clearButton: true,
      autohide: true,
      allowOneSidedRange:"ture",
      maxDate: new Date(),
    });
  });