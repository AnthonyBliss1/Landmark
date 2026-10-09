<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { Events } from "@wailsio/runtime";
  import { Landmark } from "../bindings/changeme/backend";
  import type { Movie } from "../bindings/changeme/backend";

  let movies = $state<Movie[]>([]);
  let currentMovieIndex = $state(0);
  let currentMovie = $derived(movies[currentMovieIndex] ?? null);

  let loading = $state(true);
  let trailerOverlay = $state(false);
  let error = $state("");

  async function getMovies() {
    loading = true;
    currentMovieIndex = 0;
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
    {#if trailerOverlay}
      <div class="trailer-overlay">
        <button
          class="close-trailer-overlay"
          onclick={() => {
            trailerOverlay = false;
          }}>X</button
        >
        <div class="trailer-video">
          <video class="video" src={currentMovie.Trailer.HD} controls autoplay
          ></video>
        </div>
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
          <button
            class="arrow"
            onclick={() => {
              if (currentMovieIndex === 0) {
                currentMovieIndex = movies.length - 1;
              } else {
                currentMovieIndex--;
              }
            }}
          >
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
              <dd>{currentMovie.Description}</dd>
            </dl>
            <div class="showtimes">
              {#each currentMovie.Showtimes as showtime}
                {#if !showtime.isExpired}
                  <div class="showtime">
                    {new Date(showtime.startsAt).toLocaleTimeString("en-US", {
                      hour: "numeric",
                      minute: "2-digit",
                    })}
                  </div>
                {/if}
              {/each}
            </div>
            <div class="movie-trailer">
              <button
                class="watch-trailer"
                onclick={() => {
                  trailerOverlay = true;
                }}>Watch Trailer</button
              >
            </div>
          </div>
          <button
            class="arrow"
            onclick={() => {
              if (currentMovieIndex === movies.length - 1) {
                currentMovieIndex = 0;
              } else {
                currentMovieIndex++;
              }
            }}
          >
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
    background-color: rgb(0 0 0 / 90%);
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

  .video {
    display: block;
    pointer-events: auto;
    cursor: pointer;
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
    border: 1px dashed white;
    pointer-events: auto;
    cursor: pointer;
  }

  .close-trailer-overlay:active {
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

  .landmark-location:hover {
    text-decoration: underline;
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
    border: dashed 1px;
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
    border: 1px dashed white;
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

  .movie-details dt {
    border: 1px dotted;
    padding: 0px 4px;
    align-self: start;
  }

  .movie-image {
    flex: 3;
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
    border: 1px dashed white;
  }

  .watch-trailer:active {
    border: 1px solid;
  }

  .movie-index {
    font-size: clamp(10px, 1.8vw, 18px);
  }
</style>
