<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { Events } from "@wailsio/runtime";
  import { Landmark } from "../bindings/changeme/backend";
  import type { Movie } from "../bindings/changeme/backend";

  let movies = $state<Movie[]>([]);
  let currentMovieIndex = $state(0);
  let currentMovie = $derived(movies[currentMovieIndex] ?? null);
  let currentMovieTrailer = $derived(
    currentMovie.Trailer.HD ?? currentMovie.Trailer.SD,
  ); // use HD trailer if not avail use SD
  let bookingURL = $state("");
  let timerCount = $state(5); // default set to 5 minutes

  let loading = $state(true);
  let isVideoLoading = $state(false);
  let trailerOverlay = $state(false);
  let descExpanded = $state(false);
  let error = $state("");

  let validShowtimes = $derived(
    currentMovie?.Showtimes?.filter(
      (showtime) =>
        !showtime.isExpired &&
        new Date(showtime.startsAt).getTime() > Date.now(),
    ) ?? [],
  );

  const movieDescription = document.getElementById("description");
  movieDescription?.addEventListener("click", function () {
    descExpanded = !descExpanded;
  });

  async function getMovies() {
    loading = true;
    descExpanded = false;
    currentMovieIndex = 0;
    bookingURL = "";
    error = "";

    try {
      movies = (await Landmark.GetMovies()) ?? [];
    } catch (err) {
      error = String(error);
      console.error("Failed to collect Movies:  ", err);
    } finally {
      loading = false;
    }
  }

  function bookingEventStart() {
    Events.Emit("bookingEventStart", timerCount);
  }

  function decrementMovieIndex() {
    descExpanded = false;

    if (currentMovieIndex === 0) {
      currentMovieIndex = movies.length - 1;
    } else {
      currentMovieIndex--;
    }
  }

  function incrementMovieIndex() {
    descExpanded = false;

    if (currentMovieIndex === movies.length - 1) {
      currentMovieIndex = 0;
    } else {
      currentMovieIndex++;
    }
  }

  function toggleDescription() {
    descExpanded = !descExpanded;
  }

  function videoStartLoading() {
    isVideoLoading = true;
  }

  function videoStopLoading() {
    isVideoLoading = false;
  }

  const refreshMovies = Events.On("refresh", () => {
    getMovies();
  });

  onMount(() => {
    void getMovies();
  });

  // cleanup events
  onDestroy(() => {
    refreshMovies();
  });
</script>

<main class="main-container">
  {#if loading}
    <div class="loading-container">
      <div class="loading-text">
        [ Loading movies <span class="dots"></span> ]
      </div>
    </div>
  {:else if currentMovie}
    {#if trailerOverlay && currentMovieTrailer}
      <div class="trailer-overlay">
        <button
          class="close-trailer-overlay"
          onclick={() => {
            trailerOverlay = false;
          }}>X</button
        >
        <div class="trailer-video">
          {#if isVideoLoading}
            <div class="video-spinner"></div>
          {/if}
          <video
            class="video"
            src={currentMovieTrailer}
            autoplay
            onloadstart={videoStartLoading}
            onwaiting={videoStartLoading}
            onplaying={videoStopLoading}
          ></video>
        </div>
      </div>
    {:else if bookingURL !== ""}
      <div class="booking-overlay">
        <button
          class="close-booking-overlay"
          onclick={() => {
            bookingURL = "";
          }}>X</button
        >
        <div class="booking-body">
          <p class="booking-p">
            [ Preparing booking URL <span class="dots"></span> ]
          </p>
          <p class="booking-p2">Please set a timeout for the booking session</p>
          <div class="number-input">
            <button
              class="timer-button"
              onclick={() => (timerCount = Math.max(1, timerCount - 1))}
              >−</button
            >
            <span>{String(timerCount).padStart(2, "0")} min</span>
            <button
              class="timer-button"
              onclick={() => (timerCount = Math.min(15, timerCount + 1))}
              >+</button
            >
          </div>
        </div>
        <a class="book-showtime" href={bookingURL} onclick={bookingEventStart}
          >Book Showtime</a
        >
      </div>
    {/if}
    <div class="movie-container">
      <div class="header">
        <div class="app-title">Landmark Movies</div>
        <img class="landmark-logo" src="/landmark.png" alt="landmark logo" />
        <div class="landmark-location">Pasadena</div>
      </div>

      <div class="content">
        <div class="movie-title">{currentMovie.Title}</div>
        <div class="movie-content">
          <button class="arrow" onclick={decrementMovieIndex}>
            <img src="/chevron-left.svg" alt="Previous movie" />
          </button>
          <img
            class="movie-image"
            src={currentMovie.Poster}
            alt="movie poster"
          />
          <div class="movie-info">
            <dl class="movie-details">
              <dt>Director</dt>
              <dd>{currentMovie.Director}</dd>
              <dt>Cast</dt>
              <dd>{currentMovie.Actors}</dd>
              <dt>Rating</dt>
              <dd>{currentMovie.Rating}</dd>
              <dt>Runtime</dt>
              <dd>{currentMovie.Runtime}</dd>
              <dt>Description</dt>
              <dd
                class="description"
                class:expanded={descExpanded}
                onclick={toggleDescription}
              >
                {currentMovie.Description}
              </dd>
            </dl>
            <div class="showtimes">
              {#each validShowtimes as showtime}
                <button
                  class="showtime"
                  onclick={() => {
                    bookingURL =
                      showtime?.data?.ticketing?.[0]?.urls?.[0] ?? "";
                  }}
                >
                  {new Date(showtime.startsAt).toLocaleTimeString("en-US", {
                    hour: "numeric",
                    minute: "2-digit",
                  })}
                </button>
              {:else}
                <p>No more showtimes today</p>
              {/each}
            </div>
            {#if currentMovieTrailer}
              <div class="movie-trailer">
                <button
                  class="watch-trailer"
                  onclick={() => {
                    trailerOverlay = true;
                  }}>Watch Trailer</button
                >
              </div>
            {/if}
          </div>
          <button class="arrow" onclick={incrementMovieIndex}>
            <img src="/chevron-right.svg" alt="Next movie" />
          </button>
        </div>
        <div class="movie-footer">
          <div class="movie-index">
            {currentMovieIndex + 1} / {movies.length}
          </div>
        </div>
      </div>
    </div>
  {/if}
</main>

<style>
  button {
    pointer-events: auto;
    cursor: pointer;
  }

  .main-container {
    display: flex;
    flex-direction: column;
    padding: 0 2vw 2vh;
    min-height: 100vh;
    min-width: 100vw;
    box-sizing: border-box;
    position: relative;
  }

  .loading-container {
    display: flex;
    flex: 1;
    justify-content: center;
    align-items: center;
  }

  .loading-text {
    display: inline-flex;
  }

  .dots {
    width: 3ch;
    text-align: left;
  }

  .dots::after {
    content: ".";
    animation: dots 1.2s steps(1) infinite;
  }

  @keyframes dots {
    0% {
      content: "";
    }
    25% {
      content: ".";
    }
    50% {
      content: "..";
    }
    75% {
      content: "...";
    }
    100% {
      content: "";
    }
  }

  .movie-container {
    z-index: 1;
    display: flex;
    flex-direction: column;
    flex: 1;
  }

  .trailer-overlay {
    position: absolute;
    display: flex;
    flex-direction: column;
    inset: 0;
    justify-content: center;
    align-items: center;
    background-color: rgb(0 0 0 / 95%);
    pointer-events: auto;
    z-index: 99;
  }

  .trailer-video {
    display: flex;
    justify-content: center;
    align-items: center;
    width: min(80vw, 1000px);
    max-height: 80vh;
  }

  .video-spinner {
    position: absolute;
    width: 40px;
    height: 40px;
    border: 3px solid #555;
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
  .video {
    display: block;
  }

  .close-trailer-overlay {
    position: absolute;
    top: 2vh;
    right: 2vw;
    background: black;
    color: white;
    font-size: clamp(18px, 3vw, 30px);
    cursor: pointer;
    padding: 8px 20px;
    border: 1px dotted white;
    pointer-events: auto;
    cursor: pointer;
  }

  .close-trailer-overlay:active {
    border-style: solid;
  }

  .booking-overlay {
    position: absolute;
    display: flex;
    flex-direction: column;
    inset: 0;
    justify-content: center;
    align-items: center;
    background-color: rgb(0 0 0 / 98%);
    pointer-events: auto;
    z-index: 99;
  }

  .booking-body {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    gap: 5vh;
  }

  .booking-p {
    display: inline-flex;
    font-size: clamp(18px, 2vw, 24px);
    margin-top: 0;
    margin-bottom: 0;
  }

  .booking-p2 {
    font-size: clamp(16px, 1.5vw, 20px);
    margin-top: 0;
    margin-bottom: 0;
  }

  .number-input {
    display: inline-flex;
    align-items: center;
    gap: 2vw;
  }

  .number-input span {
    display: inline-flex;
    align-items: center;
    font-size: clamp(16px, 1.5vw, 20px);
  }

  .timer-button {
    top: 2vh;
    right: 2vw;
    background: black;
    color: white;
    font-size: clamp(12px, 2vw, 24px);
    cursor: pointer;
    padding: 8px 20px;
    border: 1px dotted white;
  }

  .timer-button:active {
    border: 1px solid;
  }

  .book-showtime {
    position: absolute;
    bottom: 15vh;
    background: black;
    color: white;
    font-size: clamp(12px, 2vw, 24px);
    cursor: pointer;
    padding: 8px 20px;
    border: 1px dotted white;
    text-decoration: none;

    pointer-events: auto;
    cursor: pointer;
  }

  .book-showtime:active {
    border: 1px solid;
  }

  .close-booking-overlay {
    position: absolute;
    top: 2vh;
    right: 2vw;
    background: black;
    color: white;
    font-size: clamp(18px, 3vw, 30px);
    cursor: pointer;
    padding: 8px 20px;
    border: 1px dotted white;
    pointer-events: auto;
    cursor: pointer;
  }

  .close-booking-overlay:active {
    border-style: solid;
  }

  .header {
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    align-items: center;
    font-size: clamp(14px, 2vw, 20px);
    height: 8vh;
  }

  .app-title {
    justify-self: start;
  }

  .landmark-logo {
    justify-self: center;
    max-height: 4vh;
    width: auto;
  }

  .landmark-location {
    justify-self: end;
    pointer-events: auto;
    cursor: pointer;
    font-size: clamp(14px, 2vw, 20px);
  }

  .landmark-location:active {
    text-decoration: underline dotted 1.5px;
    text-underline-offset: 0.2em;
    pointer-events: auto;
  }

  .content {
    display: flex;
    flex-direction: column;
    flex: 1;
    padding: 2vh 2vw;
    border: solid 1px;
  }

  .movie-title {
    display: flex;
    flex: 1;
    align-items: center;
    justify-content: center;
    font-size: clamp(24px, 4vw, 40px);
    line-height: 1;
    transform: translateY(-0.2em);
  }

  .movie-content {
    display: flex;
    flex-direction: row;
    flex: 4;
    gap: 5vw;
    align-items: center;
    padding: 2vh 2vw;
    border: dotted 1px;
  }

  .movie-footer {
    display: flex;
    justify-content: center;
    padding-top: 2vh;
  }

  .arrow {
    display: flex;
    height: clamp(40px, 6rem, 60px);
    width: clamp(40px, 6rem, 60px);
    flex-shrink: 0;
    justify-content: center;
    align-items: center;
    background: black;
    box-sizing: border-box;
    border: 1px dotted white;
    pointer-events: auto;
    cursor: pointer;
  }

  .arrow img {
    display: block;
  }

  .arrow:active {
    border: 1px solid;
  }

  .movie-info {
    display: flex;
    flex: 3;
    flex-direction: column;
    justify-content: space-evenly;
    align-self: stretch;
  }

  .movie-details {
    display: grid;
    grid-template-columns: max-content 1fr;
    width: 100%;
    margin: 0;
  }

  .movie-details dt,
  .movie-details dd {
    margin: 10px;
    font-size: clamp(10px, 1.8vw, 18px);
  }

  .movie-details dd {
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 5;
    line-clamp: 5;
    overflow: hidden;

    pointer-events: auto;
    cursor: pointer;
  }

  .movie-details dd.description {
    line-height: 1.4;
    min-height: calc(5 * 1.4em);
  }

  .movie-details dd.expanded {
    -webkit-line-clamp: unset;
    line-clamp: unset;
  }

  .movie-details dt {
    border: 1px dotted;
    padding: 0px 4px;
    align-self: start;
  }

  .movie-image {
    flex: 3;

    /* need to standardize to avoid inconsistent pages */
    max-height: 67vh;
    max-width: 25vw;
  }

  .showtimes {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 10px;
    margin: 0px 10px 0;
  }

  .showtime {
    border: 1px solid;
    padding: 6px 10px;
    text-align: center;
    font-size: clamp(10px, 1.8vw, 18px);
    background: black;
  }

  .showtime:active {
    border: 1px dotted;
  }

  .showtimes p {
    grid-column: 1 / -1;
    font-size: clamp(10px, 1.8vw, 18px);
    text-align: center;
    margin: 0;
  }

  .movie-trailer {
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .watch-trailer {
    background: black;
    font-size: clamp(18px, 2vw, 20px);
    box-sizing: border-box;
    pointer-events: auto;
    cursor: pointer;
    padding: 20px 20px;
    border: 1px dotted white;
  }

  .watch-trailer:active {
    border: 1px solid;
  }

  .movie-index {
    font-size: clamp(10px, 1.8vw, 18px);
  }
</style>
