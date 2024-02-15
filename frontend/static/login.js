
    const pass1 = document.getElementById("pass1");
    const pass2 = document.getElementById("pass2");
    const email = document.getElementById("email");
    const name = document.getElementById("name");
    const signbut = document.getElementById("signbut");
    const errorMessage = document.getElementById("errorMessage");
    const email2 = document.getElementById("email2");
    const name2 = document.getElementById("name2");

    //if passwords dont match eachother if will give an error
    function checkPassword() {
        const pass1Value = pass1.value;
        const pass2Value = pass2.value;
        if (pass1Value === pass2Value) {
            pass1.style.color = '';
            pass2.style.color = '';
            pass1.style.borderColor = '';
            pass2.style.borderColor = '';
            pass1.style.borderWidth = '1px';
            pass2.style.borderWidth = '1px';
            errorMessage.innerHTML = '';
            errorMessage.style.color = '';
            return true;
        } else {
            pass2.style.color = 'red';
            pass1.style.color = 'red';
            pass1.style.borderColor = 'red';
            pass2.style.borderColor = 'red';
            pass1.style.borderWidth = '3px';
            pass2.style.borderWidth = '3px';
            errorMessage.innerHTML = 'Passwords don\'t match';
            errorMessage.style.color = 'red';
            return false;
        }
    };

    // if one or more of the imputs are empty, it will give an error
    function areInputsEmpty(inputs) {
        let asnwer = false;
        for (let i = 0; i < inputs.length; i++) {
            if (inputs[i].value.trim() === '') {
                console.log(inputs[i].value);
                inputs[i].style.borderColor = 'red';
                inputs[i].style.borderWidth = '3px';
                errorMessage.innerHTML = 'fill every input';
                errorMessage.style.color = 'red';
                asnwer = true;
            } else {
                console.log(inputs[i].value);
                inputs[i].style.borderColor = '';
                inputs[i].style.borderWidth = '';
            }
        }
        return asnwer;
    };

    // checks if inputs for sign up are ok
    function ValidateForm2() {
        const inputs = [email, name, pass1, pass2]
        if (areInputsEmpty(inputs) || !checkPassword()) {
            event.preventDefault();
            console.log('empty');
        }
    };

    // checks if inputs for log in are ok
    function ValidateForm() {
        inputs2 = [email2, name2];
        if (areInputsEmpty(inputs2)) {
            event.preventDefault();
        }
    };

