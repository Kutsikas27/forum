//@ts-nocheck

function success() {
  if (
    document.getElementById("textarea").value === "" ||
    document.getElementById("titlearea").value === "" ||
    document.getElementById("category").value === "Choose category"
  ) {
    document.getElementById("submitBtn").disabled = true;
  } else {
    document.getElementById("submitBtn").disabled = false;
  }
}
