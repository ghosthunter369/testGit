<template>
  <div class="app-container">
    <!-- 动态背景 -->
    <div class="bg-gradient"></div>
    <div class="bg-noise"></div>
    
    <router-view></router-view>
  </div>
</template>

<script setup>
</script>

<style>
@import url('https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;500;600;700&family=DM+Sans:wght@400;500&display=swap');

:root {
  --color-primary: #E8B4A0;
  --color-accent: #D4847C;
  --color-bg-dark: #1a1a2e;
  --color-bg-mid: #16213e;
  --color-text: #f8f9fa;
  --color-text-muted: rgba(248, 249, 250, 0.6);
  --color-glass: rgba(255, 255, 255, 0.08);
  --color-glass-border: rgba(255, 255, 255, 0.12);
  --radius-lg: 24px;
  --radius-md: 16px;
  --radius-sm: 12px;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'DM Sans', sans-serif;
  background: var(--color-bg-dark);
  min-height: 100vh;
  overflow: hidden;
}

.app-container {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  position: relative;
  padding: 20px;
}

/* 渐变背景 */
.bg-gradient {
  position: fixed;
  inset: 0;
  background: 
    radial-gradient(ellipse at 20% 20%, rgba(232, 180, 160, 0.15) 0%, transparent 50%),
    radial-gradient(ellipse at 80% 80%, rgba(212, 132, 124, 0.12) 0%, transparent 50%),
    radial-gradient(ellipse at 50% 50%, rgba(22, 33, 62, 0.8) 0%, var(--color-bg-dark) 70%);
  z-index: 0;
}

/* 噪点纹理 */
.bg-noise {
  position: fixed;
  inset: 0;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.8' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)'/%3E%3C/svg%3E");
  opacity: 0.03;
  pointer-events: none;
  z-index: 1;
}

/* 玻璃卡片 */
.glass-card {
  background: var(--color-glass);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid var(--color-glass-border);
  border-radius: var(--radius-lg);
  padding: 48px;
  width: 100%;
  max-width: 420px;
  position: relative;
  z-index: 10;
  box-shadow: 
    0 8px 32px rgba(0, 0, 0, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.05);
  animation: cardEntrance 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  opacity: 0;
  transform: translateY(20px);
}

@keyframes cardEntrance {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

h2 {
  font-family: 'Outfit', sans-serif;
  font-size: 28px;
  font-weight: 600;
  color: var(--color-text);
  text-align: center;
  margin-bottom: 8px;
  letter-spacing: -0.5px;
}

.subtitle {
  text-align: center;
  color: var(--color-text-muted);
  font-size: 14px;
  margin-bottom: 36px;
}

.form-group {
  margin-bottom: 20px;
  opacity: 0;
  animation: fieldEntrance 0.5s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

.form-group:nth-child(1) { animation-delay: 0.1s; }
.form-group:nth-child(2) { animation-delay: 0.15s; }
.form-group:nth-child(3) { animation-delay: 0.2s; }

@keyframes fieldEntrance {
  to {
    opacity: 1;
  }
}

label {
  display: block;
  color: var(--color-text-muted);
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 8px;
  letter-spacing: 0.3px;
  text-transform: uppercase;
}

input {
  width: 100%;
  padding: 14px 18px;
  background: rgba(0, 0, 0, 0.25);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font-size: 15px;
  font-family: 'DM Sans', sans-serif;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

input::placeholder {
  color: rgba(248, 249, 250, 0.3);
}

input:focus {
  outline: none;
  border-color: var(--color-primary);
  background: rgba(0, 0, 0, 0.35);
  box-shadow: 0 0 0 3px rgba(232, 180, 160, 0.15);
}

button {
  width: 100%;
  padding: 16px;
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-accent) 100%);
  color: var(--color-bg-dark);
  border: none;
  border-radius: var(--radius-sm);
  font-size: 15px;
  font-weight: 600;
  font-family: 'DM Sans', sans-serif;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  margin-top: 8px;
  letter-spacing: 0.5px;
  opacity: 0;
  animation: btnEntrance 0.5s cubic-bezier(0.16, 1, 0.3, 1) 0.25s forwards;
}

@keyframes btnEntrance {
  to {
    opacity: 1;
  }
}

button:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(232, 180, 160, 0.35);
}

button:active {
  transform: translateY(0);
}

.error {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #fca5a5;
  padding: 12px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  text-align: center;
  margin-bottom: 20px;
  animation: shake 0.4s ease-out;
}

.success {
  background: rgba(34, 197, 94, 0.15);
  border: 1px solid rgba(34, 197, 94, 0.3);
  color: #86efac;
  padding: 12px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  text-align: center;
  margin-bottom: 20px;
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-5px); }
  75% { transform: translateX(5px); }
}

.link {
  text-align: center;
  margin-top: 28px;
  padding-top: 24px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  opacity: 0;
  animation: linkEntrance 0.5s cubic-bezier(0.16, 1, 0.3, 1) 0.3s forwards;
}

@keyframes linkEntrance {
  to {
    opacity: 1;
  }
}

.link a {
  color: var(--color-primary);
  text-decoration: none;
  font-size: 14px;
  font-weight: 500;
  transition: color 0.2s;
}

.link a:hover {
  color: var(--color-accent);
}
</style>
