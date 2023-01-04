import {all_song_list, update_song_names, password_entered} from "/log_in.js"

const uploadsongbutton = document.getElementById("uploadsongbutton")
const uploadtrackbutton = document.getElementById("uploadtrackbutton")
const modifysongbutton = document.getElementById("modifysongbutton")
const modifytrackbutton = document.getElementById("modifytrackbutton")
const uploadsongdiv = document.getElementById("uploadsongdiv")
const uploadtrackdiv = document.getElementById("uploadtrackdiv")
const modifysongdiv = document.getElementById("modifysongdiv")
const modifytrackdiv = document.getElementById("modifytrackdiv")

const new_s_norec = document.getElementById("new_s_norec")
const new_s_rec = document.getElementById("new_s_rec")
const new_t = document.getElementById("new_t")

const song_name_choose_u = document.getElementById("choose_song_u")
const song_name_choose_s = document.getElementById("choose_song_s")
const song_name_choose_t = document.getElementById("choose_song_t")
const singing_part_choose = document.getElementById("choose_track")

const rename_s = document.getElementById("rename_s")
const rename_t = document.getElementById("rename_t")
const delete_s = document.getElementById("delete_s")
const delete_t = document.getElementById("delete_t")
const change_t = document.getElementById("change_t")
const new_track_input = document.getElementById("new_track")

let song_to_change
let track_to_change

uploadsongbutton.onclick = function(){
    uploadsongdiv.style.display = "block"
    uploadtrackdiv.style.display = "none"
    modifysongdiv.style.display = "none"
    modifytrackdiv.style.display = "none"
}
uploadtrackbutton.onclick = function(){
    uploadsongdiv.style.display = "none"
    uploadtrackdiv.style.display = "block"
    modifysongdiv.style.display = "none"
    modifytrackdiv.style.display = "none"
}
modifysongbutton.onclick = function(){
    uploadsongdiv.style.display = "none"
    uploadtrackdiv.style.display = "none"
    modifysongdiv.style.display = "block"
    modifytrackdiv.style.display = "none"
}
modifytrackbutton.onclick = function(){
    uploadsongdiv.style.display = "none"
    uploadtrackdiv.style.display = "none"
    modifysongdiv.style.display = "none"
    modifytrackdiv.style.display = "block"
}

song_name_choose_t.onchange = function (){
    singing_part_choose.selectedIndex = 0
    const prev_tracks = Array.from(singing_part_choose.options)
    prev_tracks.shift() //exclude blank/None option
    prev_tracks.forEach( track => {
        singing_part_choose.remove(track.index)
    })
    all_song_list.forEach(track => {
        if (track.song == song_name_choose_t.value) {
            const option = document.createElement("option");
            option.text = track.part;
            singing_part_choose.add(option)
        }
    })
}


new_s_norec.onclick = function (){
    upload_new_song(false)
}
new_s_rec.onclick = function (){
    upload_new_song(true)
}
const upload_new_song = function (recordable){
    const new_song_name = prompt("Please enter the new song name:")
    if (new_song_name === null){return}
    if (new_song_name.length < 3){
        alert("All song and part names must be at least 3 characters in length. No change made.")
        return
    }
    if (all_song_list.some(song => song.song == new_song_name)){
        alert("There is already a song with that name, no change made.")
        return
    }
    new_track_input.onchange = e => { 
        const new_file = e.target.files[0];
        if (new_file===undefined){
            alert("No file chosen, no change made.")
            return
        }
        if (new_file.name.split('.').pop() != "mp3"){
            alert("Sorry, only .mp3 files can be uploaded. No change made.")
            return
        }
        const new_track_name = prompt("Please enter the part name for the file chosen (e.g. Soprano, Bass, Full):") //not checking for duplicates as this is the first track
        if (new_track_name === null){return}
        if (new_track_name.length < 3){
            alert("All song and part names must be at least 3 characters in length. No change made.")
            return
        }
        do_upload(new_file,new_song_name,new_track_name,recordable).then ( () => {
            alert(`Success: New song ${new_song_name} created using file ${new_file.name} for the ${new_track_name} part. Please upload extra tracks using the 'Upload a new track for an existing song' button.`)
            update_song_names()
        })
    }
    alert(`Please choose the first mp3 file for ${new_song_name}. You will be able to upload further tracks once the song is created.`)
    new_track_input.click()
}

new_t.onclick = function (){
    song_to_change = song_name_choose_u.value
    const song_tracks = get_songs_info(song_to_change)
    if (song_tracks.length==0){
        alert("Please choose a song to add a track to")
        return
    }
    new_track_input.onchange = e => { 
        const new_file = e.target.files[0];
        if (new_file===undefined){
            alert("No file chosen, no change made.")
            return
        }
        if (new_file.name.split('.').pop() != "mp3"){
            alert("Sorry, only .mp3 files can be uploaded. No change made.")
            return
        }
        const new_track_name = prompt("Please enter the part name for the file chosen (e.g. Soprano, Bass, Full):")
        if (new_track_name === null){return}
        if (new_track_name.length < 3){
            alert("All song and part names must be at least 3 characters in length. No change made.")
            return           
        }
        if (all_song_list.some(song => song.song == song_to_change && song.part == new_track_name)){
            alert(`There is already a ${song_to_change} part with that name, no change made.`)
            return           
        }
        const recordable = song_tracks[0].recordable
        do_upload(new_file,song_to_change,new_track_name,recordable).then ( () => {
            alert(`Success: New ${new_track_name} part created for ${song_to_change} using file ${new_file.name}`)
            update_song_names()
        })
    }
    alert(`Please choose an mp3 file for ${song_to_change}`)
    new_track_input.click()
}

rename_s.onclick= function (){
    song_to_change = song_name_choose_s.value
    const song_tracks = get_songs_info(song_to_change)
    if (song_tracks.length==0){
        alert("Please choose a song to edit")
        return
    }
    const new_name = prompt("Please enter the new song name:", song_to_change)
    if (new_name === null){return}
    if (new_name.length < 3){
        alert("All song and part names must be at least 3 characters in length. No change made.")
        return
    }
    if (all_song_list.some(song => song.song == new_name)){
        alert("There is already a song with that name, no change made.")     
        return       
    }
    let counter = 0
    song_tracks.forEach( track => {
        do_rename(song_to_change,track.part,new_name,track.part).then ( () => {
            counter++
            if (counter == song_tracks.length){ //only show success once all requests finished
                alert(`Success: ${song_to_change} renamed to ${new_name}.`)
                update_song_names()
            }
        })
    })
}
delete_s.onclick= function (){
    song_to_change = song_name_choose_s.value
    const song_tracks = get_songs_info(song_to_change)
    if (song_tracks.length==0){
        alert("Please choose a song to delete")
        return
    }
    if (confirm(`This will delete ${song_to_change} and all its tracks. Are you sure?`)){
        let counter = 0
        song_tracks.forEach( track => {
            do_delete(song_to_change, track.part).then ( () => {
                counter++
                if (counter == song_tracks.length){ //only show success once all requests finished
                    alert(`Success: ${song_to_change} deleted.`)
                    update_song_names()
                }
            })
        })
    }
}

rename_t.onclick= function (){
    song_to_change = song_name_choose_t.value
    track_to_change = singing_part_choose.value
    const track = get_track_info(song_to_change,track_to_change)
    if (track === undefined){
        alert("Please choose a track to edit")
        return
    }
    const new_name = prompt("Please enter the new part name:", track_to_change)
    if (new_name === null){return}
    if (new_name.length < 3){
        alert("All song and part names must be at least 3 characters in length. No change made.")
        return
    }
    if (all_song_list.some(song => song.song == song_to_change && song.part == new_name)){
        alert(`There is already a ${song_to_change} part with that name, no change made.`)
        return           
    }
    do_rename(song_to_change,track_to_change,song_to_change,new_name).then ( () => {
        alert(`Success: ${song_to_change} part ${track_to_change} renamed to ${new_name}.`)
        update_song_names().then ( () => {
            song_name_choose_t.value = song_to_change
            song_name_choose_t.onchange()
        })
    })
}
delete_t.onclick= function (){
    song_to_change = song_name_choose_t.value
    track_to_change = singing_part_choose.value
    const track = get_track_info(song_to_change,track_to_change)
    if (track === undefined){
        alert("Please choose a track to delete")
        return
    }
    if (confirm(`Are you sure you want to delete the ${track_to_change} part for ${song_to_change}?`)){
        console.log(track)
        do_delete(song_to_change,track_to_change).then ( () => {
            alert(`Success: ${song_to_change} part ${track_to_change} deleted.`)
            update_song_names().then ( () => {
                song_name_choose_t.value = song_to_change
                song_name_choose_t.onchange()
            })
        })
    }
}
change_t.onclick= function (){
    song_to_change = song_name_choose_t.value
    track_to_change = singing_part_choose.value
    const track = get_track_info(song_to_change,track_to_change)
    if (track === undefined){
        alert("Please choose a track to edit")
        return
    }
    new_track_input.onchange = e => { 
        const new_file = e.target.files[0]; 
        if (new_file===undefined){
            alert("No file chosen, no change made.")
            return
        }
        if (new_file.name.split('.').pop() != "mp3"){
            alert("Sorry, only .mp3 files can be uploaded. No change made.")
            return
        }
        const recordable = track.recordable
        do_upload(new_file,song_to_change,track_to_change,recordable).then ( () => {
            alert(`Success: ${track_to_change} part for ${song_to_change} updated to use new file ${new_file.name}`)
        })
    }
    alert(`Please choose an mp3 file to replace the ${track_to_change} part for ${song_to_change}`)
    new_track_input.click()
}

const get_track_info = (song_chosen, track_chosen) => {
    const song = all_song_list.find(song => {
        return song.part === track_chosen &&
            song.song === song_chosen
    })
    return song
}
const get_songs_info = (song_chosen) => {
    let track_list = []
    all_song_list.forEach(track => {
        if (track.song == song_chosen) {
            track_list.push(track)
        }
    })
    return track_list
}

const do_upload = (file, song, track, recordable) => {
    return new Promise (resolve =>{
        const fd = new FormData();
        fd.append('password', password_entered)
        fd.append('new_file', file, "new_track.mp3")
        fd.append('song_name', song)
        fd.append('track_name', track)
        fd.append('recordable', recordable)
        fetch('/api/v1/uploadfile', {
            method: "post",
            body: fd
        })
        .then( res => {
            if (res.status == 200){
                resolve(true)
            }else{
                alert("Something went wrong")
                resolve(false)
            }
        })
    })
}

const do_delete = (song, track) => {
    return new Promise (resolve =>{
        const fd = new FormData();
        fd.append('password', password_entered)
        fd.append('song_name', song)
        fd.append('track_name', track)
        fetch('/api/v1/deletefile', {
            method: "post",
            body: fd
        })
        .then( res => {
            if (res.status == 200){
                resolve(true)
            }else{
                alert("Something went wrong")
                resolve(false)
            }
        })
    })
}
const do_rename = (orig_song, orig_track, new_song, new_track) => {
    return new Promise (resolve =>{
        const fd = new FormData();
        fd.append('password', password_entered)
        fd.append('orig_song_name', orig_song)
        fd.append('orig_track_name', orig_track)
        fd.append('new_song_name', new_song)
        fd.append('new_track_name', new_track)
        fetch('/api/v1/renamefile', {
            method: "post",
            body: fd
        })
        .then( res => {
            if (res.status == 200){
                resolve(true)
            }else{
                alert("Something went wrong")
                resolve(false)
            }
        })
    })
}