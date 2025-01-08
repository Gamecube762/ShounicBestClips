/**
 * fetch wrapper that errors if returned status isnt 200-299
 * @param {RequestInfo | URL} input
 * @param {RequestInit | undefined} init
 * @returns Promise
 */
const fetchWrapper = (input, init) => {
	return new Promise((res, err) => {
		fetch(input, init).then(r => {
			if (!r.ok) {
				return err(r)
			}
			res(r)
		}).catch(c => err(c))
	});
}

const expiryDate = 1736496000 * 1000;
const videoApiCfg = {
	playerVars: {
		autoplay: 0,
		controls: 1,
		disablekb: 1,
		enablejsapi: 1,
		iv_load_policy: 3,
		modestbranding: 1,
	},
}

let player1, player2;
/**
 * @type HTMLDivElement
 */
let player1Box, player2Box;
let player1Ready = false, player2Ready = false;
let selectedVote = -1; // 0 1 2
/**
 * @type HTMLDivElement
 */
let errorBox;
let videoBox, videoBoxOver;

let ___INTERNAL_ytLoadedDebounce = false
window.onYoutubeIframeAPIReady = () => {
	if (___INTERNAL_ytLoadedDebounce) return
	___INTERNAL_ytLoadedDebounce = true
	player1 = new YT.Player('player1', {
		events: { onReady: e => { player1Ready = true } },
		...videoApiCfg
	})
	player2 = new YT.Player('player2', {
		events: { onReady: e => { player2Ready = true } },
		...videoApiCfg
	})
}

const timeFormatter = (time) => {
	time /= 1000
	return `${Math.round(time/(60*60*24))}:${String(Math.round((time/(60*60)) % 24)).padStart(2, "0")}:${String(Math.round((time/60) % 60)).padStart(2, "0")}:${String(Math.round((time) % 60)).padStart(2, "0")}`
}

const voteIsOver = () => expiryDate < new Date();

const onPageLoad = () => {
	const tag = document.createElement('script');
	tag.src = "https://www.youtube.com/iframe_api";
	const firstScriptTag = document.getElementsByTagName('script')[0];
	firstScriptTag.parentNode.insertBefore(tag, firstScriptTag);

	player1Box = document.getElementById('player1Box');
	player2Box = document.getElementById('player2Box');

	errorBox = document.getElementById('errorElement');
	videoBox = document.getElementById('videos');
	videoBoxOver = document.getElementById('videos-over');

	const clock = document.getElementById('countdownTracker');
	if (voteIsOver()) {
		clock.textContent = "Voting is closed!"
		displayIsOver()
		return
	}

	selectVote(0);

	// guranteed to fire onYoutubeIframeAPIReady since api is flakey
	const checkYT = setInterval(function () {
		if(window?.YT?.loaded){
			clearInterval(checkYT);
			window.onYoutubeIframeAPIReady()
		}
	}, 100);

	const checkPlayers = setInterval(function () {
		if(player1Ready && player2Ready){
			clearInterval(checkPlayers);
			getQueue()
		}
	}, 100);

	setInterval(() => {
		clock.textContent = `${timeFormatter(expiryDate - new Date())} until voting closes!`
	}, 1000);
}

const displayIsOver = (textIfAny = "voting is over!") => {
	videoBox.classList.add('hidden')
	errorBox.classList.add('hidden')
	videoBoxOver.classList.remove('hidden')
	videoBoxOver.textContent = textIfAny;
	document.getElementsByClassName('mysides')[0].classList.add('hidden') // hacky this sux
}

const hideIsOver = () => {
	if (videoBoxOver.classList.contains('hidden')) return;
	videoBoxOver.classList.add('hidden')
	videoBox.classList.remove('hidden')
	document.getElementsByClassName('mysides')[0].classList.remove('hidden')
}

const selectVote = (what) => {
	if (what === selectedVote || voteIsOver()) return;
	selectedVote = what
	player1Box.className = (selectedVote === 1) ? 'selected' : '';
	player2Box.className = (selectedVote === 2) ? 'selected' : '';
}

const getQueue = () => {
	if (voteIsOver()) return
	return fetchWrapper("/vote/next", { headers: { 'Accept': 'application/json' } })
		.then(async (r) => {
			if (r.status == 204) {
				displayIsOver("queue is completed!")
				return
			}
			const data = await r.json();
			setVideos(data["a"], data["b"])
			hideIsOver();
		})
		.catch(async (r) => {
			try {
				if (r?.status == 420) {
					// we are done, leave
					displayIsOver()
					setError("Voting is closed!", r)
					return
				}
				const readableError = await r.json()
				setError(readableError['message'], r)
			} catch (error) {
				setError(r.message, r)
			}
		})
}

const postVote = () => {
	if (!selectedVote || voteIsOver()) return
	const form = new FormData();
	form.append("choice", (selectedVote === 1) ? player1Box.getAttribute('videoURI') : player2Box.getAttribute('videoURI'));
	fetch("/vote/submit", { method: "POST", body: new URLSearchParams(form) })
		.then(ok => {getQueue()})
		.catch(async (r) => {
			try {
				if (r?.status == 420) {
					// we are done, leave
					displayIsOver()
					setError("Voting is closed!", r)
					return
				}
				const readableError = await r.json()
				setError(readableError['message'], r)
			} catch (error) {
				setError(r.message, r)
			}
		})

	selectVote(0);
}

const setError = (what, raw) => {
	console.error("Client:", raw)
	errorBox.style = 'error'
	errorBox.textContent = what
}

function setVideos(url1, url2) {
	player1Box.setAttribute('videoURI', url1)
	player2Box.setAttribute('videoURI', url2)
	player1.loadVideoByUrl(url1)
	player1.pauseVideo()
	player2.loadVideoByUrl(url2)
	player2.pauseVideo()
}
