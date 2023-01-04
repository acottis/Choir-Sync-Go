const startdiv = document.getElementById("welcome")
const maindiv = document.getElementById("requires_password")

const password_element = document.getElementById("password_entered")
export let password_entered

const form_authenticate = document.getElementById("form_authenticate")
const admin = (form_authenticate.className == "admin")
const song_list_options = document.querySelectorAll('.song_list')

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
            update_song_names()
        }
    })
}

const authenticate = (password_passed) => {
    return new Promise (resolve =>{
        const password_send = {password: password_passed, admin: admin}

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

const get_songs = () => {
    return new Promise (resolve =>{
        const password_send = {password: password_entered, admin: admin}
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

export const update_song_names = () => {
    return new Promise (resolve =>{
        get_songs().then ( () => {
            let song_list = []
            all_song_list.forEach(track => {
                if ( !(song_list.some(song => song.name == track.song)) ) {
                    song_list.push({
                        name: track.song,
                        recordable: track.recordable,
                    })
                }
            })
            song_list_options.forEach(selector => {
                selector.selectedIndex = 0
                const prev_songs = Array.from(selector.options)
                prev_songs.shift() //exclude blank/None option
                prev_songs.forEach( track => {
                    selector.remove(track.index)
                })
                song_list.forEach(song => {
                    const option = document.createElement("option");
                    option.text = song.name;
                    if (!song.recordable){
                        option.style.color = "darkgrey"
                    }
                    selector.add(option)
                })
            })
            resolve(true)
        })
    })
}