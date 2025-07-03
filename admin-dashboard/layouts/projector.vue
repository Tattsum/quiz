<template>
  <div class="projector-layout">
    <slot />
  </div>
</template>

<style>
/* プロジェクター専用のグローバルスタイル */
.projector-layout {
  margin: 0;
  padding: 0;
  overflow: hidden;
  cursor: none; /* プロジェクター表示時はカーソルを非表示 */
}

/* 全画面表示用のリセット */
* {
  box-sizing: border-box;
}

body {
  margin: 0;
  padding: 0;
  overflow: hidden;
}

/* フルスクリーンヘルパー */
.fullscreen {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  z-index: 9999;
}
</style>