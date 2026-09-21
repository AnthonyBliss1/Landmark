<script lang="ts">
</script>

<main class="main-container">
  <div class="header">
    <div class="app-title">Landmark Movies</div>
    <div class="landmark-location">Pasadena</div>
  </div>

  <div class="content">
    <div class="movie-title">Movie Title</div>
    <div class="movie-content">
      <div class="left-arrow">{"<-"}</div>
      <div class="movie-image">image here...</div>
      <div class="movie-info">movie info here...</div>
      <div class="right-arrow">{"->"}</div>
    </div>
  </div>
</main>

<style>
  .main-container {
    display: flex;
    flex-direction: column;
    padding: 2vh 2vw;
    gap: 5vh;
    min-height: 100vh;
    min-width: 100vw;
    box-sizing: border-box;
  }

  .header {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
  }

  .landmark-location {
    pointer-events: auto;
    cursor: pointer;
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
    align-items: center;
    justify-content: center;
    min-height: 10vh;
  }

  .movie-content {
    display: flex;
    flex-direction: row;
    flex: 1;
    gap: 10vw;
    align-items: center;
    padding: 2vh 2vw;
    border: dashed 1px;
  }

  .movie-info {
    display: flex;
    flex: 3;
    border: dashed 1px;
    justify-content: center;
    align-items: center;
    align-self: stretch;
    max-width: 40vw;
  }

  .movie-image {
    display: flex;
    flex: 3;
    border: dashed 1px;
    justify-content: center;
    align-items: center;
    align-self: stretch;
    max-width: 40vw;
  }
</style>
