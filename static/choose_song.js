import {password_entered} from "/log_in.js"

const want_sim = document.getElementById("want_sim")
const want_adv = document.getElementById("want_adv")
const want_rec = document.getElementById("want_rec")
const song_name_choose = document.getElementById("choose_song")
const singing_part_choose = document.getElementById("choose_track")
const singing_part_chooseL = document.getElementById("choose_trackL")
const singing_part_chooseR = document.getElementById("choose_trackR")
const backing_player = document.getElementById("backing_player")
const dual_player1 = document.getElementById("dual_player1")
const dual_player2 = document.getElementById("dual_player2")

const yes_record_divs = document.querySelectorAll('.yes_record')
const no_record_divs = document.querySelectorAll('.no_record')
const simdivs = document.querySelectorAll('.sim')
const advdivs = document.querySelectorAll('.adv')
const safaridivs = document.querySelectorAll('.safari')
const recdivs = document.querySelectorAll('.rec')

let all_song_list = []
let recordable = true
let on_rec_tab
let is_safari
let secret

let audioCtx
let panNode1
let panNode2
let switch_channels = false

export let song_name
export let singing_part
export let backing_track_file
export let song_is_chosen = false
export let backing_audio_playing = false
export let mimetype_chosen

if (navigator.userAgent.indexOf('Safari') != -1 && navigator.userAgent.indexOf('Chrome') == -1) {
    is_safari = true
}

export const set_up_songs = async () => {
    get_songs().then ( password_correct => {
        if (password_correct) {
            change_song_names()
        }
    })
    
    if(MediaRecorder.isTypeSupported("audio/webm")){
        mimetype_chosen = "mimeType: audio/webm"
    }
    else if(MediaRecorder.isTypeSupported("audio/mpeg")){
        mimetype_chosen = "mimeType: audio/mpeg"
    }
    else{
        mimetype_chosen = ""
    }
}

const change_song_names = () => {
    let song_list = []
    all_song_list.forEach(track => {
        if ( !(song_list.some(song => song.name == track.song)) ) {
            song_list.push({
                name: track.song,
                recordable: track.recordable,
            })
        }
    })
    song_list.forEach(song => {
        const option = document.createElement("option");
        option.text = song.name;
        if (!song.recordable){
            option.style.color = "darkgrey"
        }
        song_name_choose.add(option)
    })
}
 
const change_track_names = () => {
    change_track_names_X(singing_part_choose)
    change_track_names_X(singing_part_chooseL)
    change_track_names_X(singing_part_chooseR)
}
const change_track_names_X = (singing_part_chooseX) => {
    singing_part_chooseX.selectedIndex = 0
    const prev_tracks = Array.from(singing_part_chooseX.options)
    prev_tracks.shift() //exclude blank/None option
    prev_tracks.forEach( track => {
        singing_part_chooseX.remove(track.index)
    })
    all_song_list.forEach(track => {
        if (track.song == song_name) {
            const option = document.createElement("option");
            option.text = track.part;
            singing_part_chooseX.add(option)
            recordable = track.recordable
            secret = track.secret
        }
    })
}

const get_songs = () => {
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

song_name_choose.onchange = function (){
    song_name = song_name_choose.value
    song_is_chosen = false
    backing_player.removeAttribute("src")
    dual_player1.removeAttribute("src")
    dual_player2.removeAttribute("src")
    change_track_names()
    if(!recordable){
        if(on_rec_tab){
            yes_record_divs.forEach(div => {
                div.style.display = "none"
            });
        }
        no_record_divs.forEach(div => {
            div.style.display = "inline"
        });
    }
    else{
        if(on_rec_tab){
            yes_record_divs.forEach(div => {
                div.style.display = "block"
            });
        }
        no_record_divs.forEach(div => {
            div.style.display = "none"
        });
    }
}

const change_track = (player) => {

    // Get the currently selected song
    const song = all_song_list.find(song => {
        // Return the first match of the song name and part matching
        return song.part === singing_part_choose.value &&
            song.song === song_name_choose.value
    })
    /// TODO - Handle undefined if song not found!

    switch(player){
        case "main":
            backing_player.setAttribute("src", song.url)
            break;
        case "L":
            if(song.url == ""){
                dual_player1.setAttribute("src", dual_player2.src)
                dual_player2.muted=true
                switch_channels=true
            }
            else{
                dual_player1.setAttribute("src", song.url)
                switch_channels=false
            }
            dual_player2.pause()
            break;
        case "R":
            dual_player2.setAttribute("src", song.url)
            dual_player1.pause()
            dual_player1.currentTime=0
            if(dual_player1.src==""){
                change_track("blank", "L")
            }
    }
}

singing_part_choose.onchange = function (){
    change_track("main")
}
singing_part_chooseL.onchange = function (){
    change_track("L")
}
singing_part_chooseR.onchange = function (){
    change_track("R")
}

backing_player.addEventListener("play", event => {
    backing_audio_playing = true
});
backing_player.addEventListener("pause", event => {
    backing_audio_playing = false
});
backing_player.addEventListener("ended", event => {
    backing_audio_playing = false
});

dual_player1.addEventListener("play", event => {

    if (!audioCtx) {
      audioCtx = new AudioContext();
        let source1 = new MediaElementAudioSourceNode(audioCtx, {
            mediaElement: dual_player1,
        });
        let source2 = new MediaElementAudioSourceNode(audioCtx, {
            mediaElement: dual_player2,
        });
        panNode1 = new StereoPannerNode(audioCtx);
        panNode2 = new StereoPannerNode(audioCtx);
        panNode1.pan.value = -1;
        panNode2.pan.value = 1;
        source1.connect(panNode1);
        source2.connect(panNode2);
        panNode1.connect(audioCtx.destination);
        panNode2.connect(audioCtx.destination);
    }

    dual_player2.currentTime = dual_player1.currentTime
    dual_player2.volume = dual_player1.volume
    dual_player2.playbackRate = dual_player1.playbackRate
    if(switch_channels){
        panNode1.pan.value = 1;
        panNode2.pan.value = -1;
    }
    else{
        panNode1.pan.value = -1;
        panNode2.pan.value = 1;
        dual_player2.muted = dual_player1.muted
    }
    dual_player2.play()
    backing_audio_playing = true
});
dual_player1.addEventListener("pause", event => {
    dual_player2.pause()
    backing_audio_playing = false
});
dual_player1.addEventListener("ended", event => {
    backing_audio_playing = false
});

document.addEventListener('play', function(e){
    if(e.target.id != "dual_player2"){
        var audios = document.getElementsByTagName('audio');
        for(var i = 0, len = audios.length; i < len;i++){
            if(audios[i] != e.target && !((audios[i].id == "dual_player1") && (e.target.id == "dual_player2"))){
                audios[i].pause();
            }
        }
    }
}, true);

want_sim.onclick = function(){
    pause_all()
    advdivs.forEach(div => {
        div.style.display = "none"
    });
    recdivs.forEach(div => {
        div.style.display = "none"
    });
    simdivs.forEach(div => {
        div.style.display = "block"
    });
    on_rec_tab = false
}
want_adv.onclick = function(){
    pause_all()
    simdivs.forEach(div => {
        div.style.display = "none"
    });
    recdivs.forEach(div => {
        div.style.display = "none"
    });
    if(is_safari){
        advdivs.forEach(div => {
            div.style.display = "none"
        });
        safaridivs.forEach(div => {
            div.style.display = "block"
        });
    }
    else{
        advdivs.forEach(div => {
            div.style.display = "block"
        });
        safaridivs.forEach(div => {
            div.style.display = "none"
        });
    }
    on_rec_tab = false
}
want_rec.onclick = function(){
    pause_all()
    advdivs.forEach(div => {
        div.style.display = "none"
    });
    simdivs.forEach(div => {
        div.style.display = "none"
    });
    recdivs.forEach(div => {
        div.style.display = "block"
    });
    on_rec_tab = true
    if(!recordable){
        yes_record_divs.forEach(div => {
            div.style.display = "none"
        });
    }
}

const pause_all = () => {
    var audios = document.getElementsByTagName('audio');
    for(var i = 0, len = audios.length; i < len;i++){
        audios[i].pause();
    }
}