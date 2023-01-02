const startdiv = document.getElementById("welcome")
const maindiv = document.getElementById("requires_password")

const password_element = document.getElementById("password_entered")
export let password_entered

const form_authenticate = document.getElementById("form_authenticate")
const admin = (form_authenticate.className == "admin")

export let all_song_list = []

const log_in = () => {
    password_entered = password_element.value
    show_page()
}

const show_page = async () => {
    authenticate(password_entered).then ( password_correct => {
        if (password_correct) {
            startdiv.style.display = "none"
            maindiv.style.display = "block"
            if (admin){
                get_songs()
            }
        }
    })
}

const authenticate = (password_passed) => {
    return new Promise (resolve =>{
        const password_send = {password: password_passed}

        fetch(`/api/v1/auth`, {
            //also pass whether we're signing in to admin or normal
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
    event.preventDefault(); //stop page refresh on form submit
    return false
})

export const get_songs = () => {
    return new Promise (resolve =>{
        const password_send = {password: password_entered}
        fetch(`/api/v1/getsongs`, {
            method: "post",
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(password_send)
        })
        .then( res => {
            if (res.status == 200){
                res.json().then( res => {
                    all_song_list = res
                    resolve(true)
                })
            }else{
                alert(
                    "Server could not fetch songs, please contact administrator"
                )
                resolve(false)
            }
        })
    })
}