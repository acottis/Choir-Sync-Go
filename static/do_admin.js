import {all_song_list, update_song_names} from "/log_in.js"

const uploadbutton = document.getElementById("uploadbutton")
const modifysongbutton = document.getElementById("modifysongbutton")
const modifytrackbutton = document.getElementById("modifytrackbutton")
const uploaddiv = document.getElementById("uploaddiv")
const modifysongdiv = document.getElementById("modifysongdiv")
const modifytrackdiv = document.getElementById("modifytrackdiv")

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

uploadbutton.onclick = function(){
    uploaddiv.style.display = "block"
    modifysongdiv.style.display = "none"
    modifytrackdiv.style.display = "none"
}
modifysongbutton.onclick = function(){
    uploaddiv.style.display = "none"
    modifysongdiv.style.display = "block"
    modifytrackdiv.style.display = "none"
}
modifytrackbutton.onclick = function(){
    uploaddiv.style.display = "none"
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

rename_s.onclick= function (){
    song_to_change = song_name_choose_s.value
    const test_songs = get_songs_info(song_to_change)
    if (test_songs.length==0){
        alert("Please choose a song to edit")
    } else {
        const new_name = prompt("Please enter the new song name:", song_to_change)
        if (all_song_list.some(song => song.song == new_name)){
            alert("There is already a song with that name, no change made.")            
        } else {
            console.log(test_songs)
            console.log(new_name)
            //DO: change the song name for all these tracks
            update_song_names()
        }
    }
}
delete_s.onclick= function (){
    song_to_change = song_name_choose_s.value
    const test_songs = get_songs_info(song_to_change)
    if (test_songs.length==0){
        alert("Please choose a song to delete")
    } else {
        if (confirm(`This will delete ${song_to_change} and all its tracks. Are you sure?`)){
            console.log(test_songs)
            //DO: delete all these tracks
            update_song_names()
        }
    }
}
rename_t.onclick= function (){
    song_to_change = song_name_choose_t.value
    track_to_change = singing_part_choose.value
    const track = get_track_info(song_to_change,track_to_change)
    if (track === undefined){
        alert("Please choose a track to edit")
    } else {
        const new_name = prompt("Please enter the new part name:", track_to_change)
        if (all_song_list.some(song => song.song == song_to_change && song.part == new_name)){
            alert(`There is already a ${song_to_change} part with that name, no change made.`)            
        } else {
            console.log(track)
            console.log(new_name)
            //DO: change this track part name
            update_song_names().then ( () => {
                song_name_choose_t.value = song_to_change
                song_name_choose_t.onchange()
            })
        }
    }
}
delete_t.onclick= function (){
    song_to_change = song_name_choose_t.value
    track_to_change = singing_part_choose.value
    const track = get_track_info(song_to_change,track_to_change)
    if (track === undefined){
        alert("Please choose a track to delete")
    } else {
        if (confirm(`Are you sure you want to delete the ${track_to_change} part for ${song_to_change}?`)){
            console.log(track)
            //DO: delete this track
            update_song_names().then ( () => {
                song_name_choose_t.value = song_to_change
                song_name_choose_t.onchange()
            })
        }
    }
}
change_t.onclick= function (){
    song_to_change = song_name_choose_t.value
    track_to_change = singing_part_choose.value
    const track = get_track_info(song_to_change,track_to_change)
    let new_file
    if (track === undefined){
        alert("Please choose a track to edit")
    } else {
        new_track_input.onchange = e => { 
            new_file = e.target.files[0]; 
            console.log(track)
            console.log(new_file)
            //DO: change the file for this track (or delete and add)
        }
        alert(`Please choose an mp3 file to replace the ${track_to_change} part for ${song_to_change}`)
        new_track_input.click()
    }
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