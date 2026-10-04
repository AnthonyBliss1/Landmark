<script lang="ts">
  import { onMount } from "svelte";
  import { Landmark } from "../bindings/changeme/backend";
  import type { Movie } from "../bindings/changeme/backend";

  let movies = $state<Movie[]>([]);
  let currentMovieIndex = $state(0);
  let currentMovie = $derived(movies[currentMovieIndex] ?? null);

  let loading = $state(true);
  let error = $state("");

  async function getMovies() {
    loading = true;
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

  onMount(() => {
    void getMovies();
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
          {"<-"}
        </button>
        <img class="movie-image" src={currentMovie.Poster} alt="movie poster" />
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
          {"->"}
        </button>
      </div>
      <div class="movie-footer">
        <div class="movie-index">{currentMovieIndex + 1} / {movies.length}</div>
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
    width: 40px;
    flex-shrink: 0;

    background: black;
    font-size: clamp(18px, 2.5vw, 25px);
    box-sizing: border-box;
    pointer-events: auto;
    cursor: pointer;
    padding: 2px;
    border: 1px solid transparent;
  }

  .arrow:hover {
    border: 1px solid;
  }

  .movie-info {
    display: flex;
    flex: 3;
    flex-direction: column;
    justify-content: center;
    align-items: stretch;
    align-self: stretch;
  }

  .movie-details {
    display: grid;
    grid-template-columns: max-content 1fr;
    width: 100%;
  }

  .movie-details dt,
  .movie-details dd {
    margin: 10px;
    font-size: clamp(10px, 1.8vw, 18px);
  }

  .movie-details dt {
    border: 1px dotted;
    padding: 4px;
    align-self: start;
  }

  .movie-image {
    display: flex;
    flex: 3;
    justify-content: center;
    align-items: center;
    max-width: 25vw;
  }

  .showtimes {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 10px;
    margin: 20px 10px 0;
  }

  .showtime {
    border: 1px solid;
    padding: 6px 10px;
    text-align: center;
    font-size: clamp(10px, 1.8vw, 18px);
  }

  .movie-index {
    font-size: clamp(10px, 1.8vw, 18px);
  }
</style>
