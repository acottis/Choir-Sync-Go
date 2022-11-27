import {set_up_songs} from "/choose_song.js"

const startdiv = document.getElementById("welcome")
const maindiv = document.getElementById("requires_password")

const password_element = document.getElementById("password_entered")
export let password_entered

const form_authenticate = document.getElementById("form_authenticate")

const log_in = () => {
    password_entered = password_element.value
    show_page()
}

const show_page = async () => {
    authenticate(password_entered).then ( password_correct => {
        if (password_correct) {
            startdiv.style.display = "none"
            maindiv.style.display = "block"
            set_up_songs()
        }
    })
}

const authenticate = (password_passed) => {
    return new Promise (resolve =>{
        const password_send = {password: password_passed}

        fetch(`/api/v1/auth`, {
            method: "post",
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(password_send)
        })
        .then( res => {
            if (res.status == 200){
                resolve(true)
            }
            else {
                alert("Sorry, the password is not correct")
                resolve(false)
            }
        })
    })
}

form_authenticate.addEventListener("submit", event  => {
    log_in()

    //stop page refresh on form submit
    event.preventDefault();
    return false
})