# AZSUBAY Landing Page Design System (Pure HTML/CSS)

> Extracted from `index.html`, splash screen, and all homepage sections.
> Framework-free — ready to drop into any project.

---

## 1. SPLASH SCREEN

Paste this directly in your `.html` file, inside `<body>`, BEFORE your main content container:

```html
<style>
  #app-splash {
    position: fixed;
    inset: 0;
    z-index: 9999;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background: #ffffff;
    transition: opacity 0.45s ease;
    gap: 20px;
  }
  #app-splash.fade {
    opacity: 0;
    pointer-events: none;
  }
  #app-splash img {
    width: 128px;
    height: 128px;
    object-fit: contain;
    animation: splash-pop 0.6s ease-out both, splash-bob 2.4s ease-in-out 0.6s infinite;
  }
  #app-splash .brand {
    font-family: 'Poppins', 'Segoe UI', system-ui, sans-serif;
    font-weight: 800;
    font-size: 28px;
    letter-spacing: 0.5px;
    background: linear-gradient(90deg, #1f2937 0%, #1f2937 35%, #ffffff 50%, #1f2937 65%, #1f2937 100%);
    background-size: 200% 100%;
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
    animation: splash-shine 1.8s linear infinite;
  }
  #app-splash .dots {
    display: flex;
    gap: 6px;
  }
  #app-splash .dots span {
    width: 6px;
    height: 6px;
    border-radius: 9999px;
    background: #1f2937;
    opacity: 0.25;
    animation: splash-dot 1.2s ease-in-out infinite;
  }
  #app-splash .dots span:nth-child(2) { animation-delay: 0.15s; }
  #app-splash .dots span:nth-child(3) { animation-delay: 0.3s; }

  @keyframes splash-shine {
    0%   { background-position: 200% 0; }
    100% { background-position: -200% 0; }
  }
  @keyframes splash-pop {
    from { transform: scale(0.6); opacity: 0; }
    to   { transform: scale(1);   opacity: 1; }
  }
  @keyframes splash-bob {
    0%, 100% { transform: translateY(0); }
    50%      { transform: translateY(-6px); }
  }
  @keyframes splash-dot {
    0%, 100% { opacity: 0.25; transform: translateY(0); }
    50%      { opacity: 1;    transform: translateY(-4px); }
  }
</style>

<div id="app-splash">
  <img src="/logo.png" alt="Logo" />
  <div class="brand">Brand Name</div>
  <div class="dots"><span></span><span></span><span></span></div>
</div>

<script>
  window.addEventListener('load', function () {
    setTimeout(function () {
      var s = document.getElementById('app-splash');
      if (s) {
        s.classList.add('fade');
        setTimeout(function () { s.remove(); }, 500);
      }
    }, 600);
  });
</script>
```

**How it works:** Logo pops in, bobs gently, brand name has a shine animation, three dots pulse. After 600ms the whole thing fades out and gets removed from DOM.

---

## 2. GLOBAL STYLESHEET (CSS Reset + Tokens + Base Classes)

```css
/* ===== FONTS ===== */
@import url('https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600;700;800&family=Inter:wght@300;400;500;600;700&display=swap');

/* ===== CSS VARIABLES (Design Tokens) ===== */
:root {
  --color-bg:        hsl(0, 0%, 100%);   /* #ffffff */
  --color-fg:        hsl(0, 0%, 8%);     /* #141414 (near black) */
  --color-muted:     hsl(0, 0%, 94%);    /* #f0f0f0 */
  --color-muted-fg:  hsl(0, 0%, 45%);    /* #737373 (gray text) */
  --color-secondary: hsl(0, 0%, 96%);    /* #f5f5f5 */
  --color-elevated:  hsl(0, 0%, 97%);    /* #f7f7f7 */
  --color-card:      hsl(0, 0%, 98%);    /* #fafafa */
  --color-border:    hsl(0, 0%, 90%);    /* #e5e5e5 */
  --color-input:     hsl(0, 0%, 90%);
  --color-ring:      hsl(0, 0%, 8%);
  --color-primary:   hsl(0, 0%, 8%);
  --color-primary-fg:hsl(0, 0%, 100%);
  --color-destructive:hsl(0, 84%, 60%);
  --color-brand-red: hsl(0, 80%, 50%);
  --color-brand-blue:hsl(217, 91%, 55%);
  --radius:          0.75rem;
  --font-heading:    'Poppins', 'Segoe UI', system-ui, sans-serif;
  --font-body:       'Inter', 'Segoe UI', system-ui, sans-serif;
}

/* ===== RESET ===== */
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
html { -webkit-font-smoothing: antialiased; -moz-osx-font-smoothing: grayscale; }
body {
  font-family: var(--font-body);
  background: var(--color-bg);
  color: var(--color-fg);
  line-height: 1.6;
}
h1, h2, h3, h4, h5, h6 { font-family: var(--font-heading); }

/* ===== UTILITY CLASSES ===== */
.container {
  width: 100%;
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 1.5rem;
}
.sr-only {
  position: absolute; width: 1px; height: 1px; overflow: hidden;
  clip: rect(0,0,0,0); white-space: nowrap; border: 0;
}
.scrollbar-hide::-webkit-scrollbar { display: none; }
.scrollbar-hide { -ms-overflow-style: none; scrollbar-width: none; }

/* ===== BUTTON BASE ===== */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-family: var(--font-body);
  font-weight: 600;
  border: 2px solid transparent;
  cursor: pointer;
  text-decoration: none;
  transition: all 0.15s ease;
  white-space: nowrap;
}
.btn-primary {
  background: var(--color-fg);
  color: var(--color-primary-fg);
  border-radius: 9999px;
  padding: 0 2rem;
  height: 3rem;
  font-size: 1rem;
}
.btn-primary:hover { opacity: 0.9; }
.btn-outline {
  background: var(--color-bg);
  color: var(--color-fg);
  border-color: var(--color-border);
  border-radius: 9999px;
  padding: 0 2rem;
  height: 3rem;
  font-size: 1rem;
}
.btn-outline:hover { background: var(--color-secondary); }
.btn-icon {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

/* ===== BADGE ===== */
.badge {
  display: inline-flex;
  align-items: center;
  padding: 0.125rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.625rem; /* 10px */
  font-weight: 500;
  background: var(--color-secondary);
  color: var(--color-fg);
  line-height: 1.5;
}

/* ===== SECTION PADDING ===== */
.section { padding: 3rem 0; }          /* py-12 */
.section-hero { padding: 2.5rem 0; }   /* py-10 */

@media (min-width: 768px) {
  .section { padding: 5rem 0; }        /* md:py-20 */
  .section-hero { padding: 5rem 0; }   /* md:py-20 */
}
@media (min-width: 1024px) {
  .section-hero { padding: 7rem 0; }   /* lg:py-28 */
}
```

---

## 3. PAGE LAYOUT (Full Structure)

The homepage follows a **stacked section** pattern — header → sections → footer.

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Your Site</title>
  <link rel="icon" href="/favicon.png" />
  <link rel="stylesheet" href="/style.css" />
</head>
<body>

  <!-- SPLASH SCREEN (paste Section 1 here) -->

  <div id="page" style="min-height:100vh; display:flex; flex-direction:column;">

    <!-- HEADER -->
    <header id="header" style="position:sticky; top:0; z-index:50; background:var(--color-bg); border-bottom:1px solid var(--color-border);">
      <div class="container" style="display:flex; align-items:center; justify-content:space-between; height:4rem;">
        <a href="/" style="font-family:var(--font-heading); font-weight:800; font-size:1.25rem; text-decoration:none; color:var(--color-fg);">
          LOGO
        </a>
        <nav style="display:none;">
          <!-- Desktop nav links -->
          <a href="/products" style="text-decoration:none; color:var(--color-muted-fg); font-size:0.875rem; padding:0 0.75rem;">Products</a>
          <a href="/about"  style="text-decoration:none; color:var(--color-muted-fg); font-size:0.875rem; padding:0 0.75rem;">About</a>
          <a href="/contact" style="text-decoration:none; color:var(--color-muted-fg); font-size:0.875rem; padding:0 0.75rem;">Contact</a>
        </nav>
        <!-- Mobile hamburger button here -->
      </div>
    </header>

    <!-- MAIN CONTENT -->
    <main style="flex:1;">
      <!-- Paste Hero section here (Section 4) -->
      <!-- Paste Category Grid here (Section 5) -->
      <!-- Paste Feature Cards here (Section 6) -->
      <!-- Paste CTA Banner here (Section 7) -->
    </main>

    <!-- FOOTER -->
    <footer id="footer" style="background:var(--color-secondary); border-top:1px solid var(--color-border);">
      <div class="container" style="padding-top:3rem; padding-bottom:3rem; text-align:center;">
        <p style="font-size:0.875rem; color:var(--color-muted-fg);">
          &copy; 2025 Your Brand. All rights reserved.
        </p>
      </div>
    </footer>

  </div>
</body>
</html>
```

**Key rules:**
- `min-height:100vh; display:flex; flex-direction:column` on wrapper pushes footer to bottom on short pages
- `flex:1` on `<main>` fills remaining space between header and footer
- Header is `position:sticky; top:0` with bottom border

---

## 4. SECTION RECIPE: Hero Banner

```html
<!-- HERO SECTION -->
<section style="background: var(--color-elevated);" class="section-hero">
  <div class="container">

    <div class="hero-grid">
      <!-- LEFT: Text -->
      <div class="hero-text animate-fade-up">
        <h1 class="hero-headline">
          Turn ideas into income.
          <span class="hero-headline-muted">Scale without limits.</span>
        </h1>
        <p class="hero-desc">
          Tools, systems, and products to help you start, grow, and expand.
        </p>
        <a href="/products" class="btn btn-primary">Explore Solutions</a>

        <!-- Stats row -->
        <div class="hero-stats">
          <div class="hero-stat">
            <span class="hero-stat-label">Trusted Platform</span>
            <span class="hero-stat-sub">Verified Solutions</span>
          </div>
          <div class="hero-stat">
            <span class="hero-stat-label">50+</span>
            <span class="hero-stat-sub">Business Solutions</span>
          </div>
        </div>
      </div>

      <!-- RIGHT: Image -->
      <div class="hero-image-wrap animate-fade-scale">
        <img src="/hero.png" alt="Hero" class="hero-image" />
      </div>
    </div>

  </div>
</section>

<style>
/* Hero grid */
.hero-grid {
  display: grid;
  gap: 1.5rem;
  align-items: center;
}
@media (min-width: 768px) {
  .hero-grid { grid-template-columns: 1fr 1fr; gap: 3rem; }
  .hero-image-wrap { order: 2; } /* image on right */
}

/* Hero text */
.hero-text { text-align: left; }
.hero-headline {
  font-family: var(--font-heading);
  font-size: 1.875rem;      /* 30px mobile */
  font-weight: 700;
  line-height: 1.1;
  letter-spacing: -0.025em;
  margin-bottom: 1rem;
}
.hero-headline-muted { color: var(--color-muted-fg); }
.hero-desc {
  font-size: 1rem;
  color: var(--color-muted-fg);
  margin-bottom: 2rem;
  max-width: 28rem;
}
@media (min-width: 768px) {
  .hero-headline { font-size: 3rem; }    /* 48px */
  .hero-desc { font-size: 1.125rem; }
}
@media (min-width: 1024px) {
  .hero-headline { font-size: 3.5rem; }  /* 56px */
}

/* Hero stats */
.hero-stats {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin-top: 1.5rem;
  max-width: 28rem;
}
@media (min-width: 768px) { .hero-stats { margin-top: 3.5rem; } }
.hero-stat {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  padding: 0.75rem;
  border-radius: 1rem;       /* rounded-2xl */
  background: var(--color-bg);
}
.hero-stat-label {
  font-family: var(--font-heading);
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1.25;
}
.hero-stat-sub {
  font-size: 0.75rem;
  color: var(--color-muted-fg);
}

/* Hero image */
.hero-image-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 1rem;
}
@media (min-width: 768px) { .hero-image-wrap { margin-top: 0; } }
.hero-image {
  width: 100%;
  max-width: 220px;
  height: auto;
  object-fit: contain;
  filter: drop-shadow(0 20px 13px rgba(0,0,0,0.03)) drop-shadow(0 8px 5px rgba(0,0,0,0.08));
}
@media (min-width: 768px) { .hero-image { max-width: 28rem; } }
</style>
```

**Rules:**
- Background: `var(--color-elevated)` (97% gray)
- 2-column grid at 768px+, stacks on mobile
- Headline: Poppins, bold, tight tracking; muted span for secondary emphasis
- CTA: pill-shaped (`border-radius: 9999px`), 3rem tall
- Stat cards: white background, 1rem border-radius, 2×2 grid

---

## 5. SECTION RECIPE: Category Grid (Circles)

Mobile: horizontal scroll | Desktop: 7-column grid

```html
<!-- CATEGORY GRID SECTION -->
<section class="section">
  <div class="container">

    <div class="section-header">
      <h2 class="section-title">Browse Categories</h2>
      <p class="section-subtitle">Find what you need</p>
    </div>

    <!-- Scroll container (mobile) / Grid (desktop) -->
    <div class="cat-scroll">
      <!-- Repeat for each category -->
      <a href="/products?cat=courses" class="cat-item">
        <div class="cat-circle group">
          <img src="cat-1.jpg" alt="Courses" class="cat-img cat-img-default" loading="lazy" />
          <img src="cat-1-hover.jpg" alt="" class="cat-img cat-img-hover" loading="lazy" aria-hidden="true" />
        </div>
        <span class="cat-label">Courses</span>
      </a>
      <a href="/products?cat=books" class="cat-item">
        <div class="cat-circle group">
          <img src="cat-2.jpg" alt="Books" class="cat-img cat-img-default" loading="lazy" />
          <img src="cat-2-hover.jpg" alt="" class="cat-img cat-img-hover" loading="lazy" aria-hidden="true" />
        </div>
        <span class="cat-label">Books</span>
      </a>
      <!-- ... more categories ... -->
    </div>

  </div>
</section>

<style>
/* Section header */
.section-header { text-align: center; margin-bottom: 2rem; }
.section-title {
  font-family: var(--font-heading);
  font-size: 1.125rem;     /* text-lg */
  font-weight: 600;
  margin-bottom: 0.25rem;
}
.section-subtitle {
  font-size: 0.875rem;      /* text-sm */
  color: var(--color-muted-fg);
}
@media (min-width: 768px) {
  .section-title { font-size: 1.25rem; }  /* md:text-xl */
}

/* Category scroll / grid */
.cat-scroll {
  display: flex;
  gap: 1rem;
  overflow-x: auto;
  padding-bottom: 0.5rem;
  scroll-snap-type: x mandatory;
  scroll-behavior: smooth;
}
.cat-scroll::-webkit-scrollbar { display: none; }  /* hide scrollbar */
.cat-scroll { -ms-overflow-style: none; scrollbar-width: none; }

@media (min-width: 768px) {
  .cat-scroll {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    overflow-x: visible;
    padding-bottom: 0;
  }
}

/* Category item */
.cat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  min-width: 90px;
  scroll-snap-align: center;
  text-decoration: none;
}
.cat-label {
  font-size: 0.75rem;       /* text-xs */
  font-weight: 500;
  color: var(--color-fg);
}

/* Circle with double-image hover swap */
.cat-circle {
  position: relative;
  width: 5rem;              /* 80px */
  height: 5rem;
  border-radius: 9999px;
  overflow: hidden;
  border: 2px solid var(--color-border);
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
  transition: all 0.5s ease;
}
@media (min-width: 768px) {
  .cat-circle { width: 6rem; height: 6rem; }  /* 96px */
}
.cat-circle:hover {
  box-shadow: 0 10px 15px -3px rgba(0,0,0,0.1);
  box-shadow: 0 0 0 4px hsl(0, 0%, 96%);
}
.cat-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: all 0.7s ease;
}
.cat-img-default {
  /* visible by default */
}
.cat-circle:hover .cat-img-default {
  opacity: 0;
  transform: scale(1.1);
}
.cat-img-hover {
  opacity: 0;
  transform: scale(1.1);
}
.cat-circle:hover .cat-img-hover {
  opacity: 1;
  transform: scale(1);
}
</style>
```

**Rules:**
- `display: flex` on mobile → horizontal scroll with `scroll-snap-type: x mandatory`
- `display: grid; grid-template-columns: repeat(7, 1fr)` at 768px+
- Circle images: `border-radius: 9999px`, 2px border, overflow hidden
- Hover: base image fades out + scales up, hover image fades in + scales to normal
- Items are `<a>` links wrapping the circle + label

---

## 6. SECTION RECIPE: Feature Cards / Pricing Bundles

```html
<!-- FEATURE CARDS SECTION -->
<section class="section">
  <div class="container">

    <div class="section-header" style="text-align:left;">
      <h2 class="section-title" style="font-size:1.5rem; font-weight:700; letter-spacing:-0.025em;">Featured solutions</h2>
      <p class="section-subtitle" style="margin-top:0.25rem;">Curated bundles for every level</p>
    </div>

    <div class="cards-grid">
      <!-- Card 1 -->
      <a href="/products" class="card">
        <span class="badge card-badge">Popular</span>
        <div class="card-icon-box">
          <!-- SVG icon here or empty div -->
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/></svg>
        </div>
        <h3 class="card-title">Beginner Pack</h3>
        <p class="card-desc">Go from zero to your first deployment in 30 days.</p>
        <div class="card-footer">
          <span class="card-price">$29</span>
          <span class="card-action">View →</span>
        </div>
      </a>

      <!-- Card 2 -->
      <a href="/products" class="card">
        <div class="card-icon-box">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2L2 7l10 5 10-5-10-5z"/></svg>
        </div>
        <h3 class="card-title">Advanced Pack</h3>
        <p class="card-desc">Master networking, DevOps, and cloud infrastructure.</p>
        <div class="card-footer">
          <span class="card-price">$79</span>
          <span class="card-action">View →</span>
        </div>
      </a>

      <!-- Card 3 -->
      <a href="/products" class="card">
        <span class="badge card-badge">Best Value</span>
        <div class="card-icon-box">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="2" width="20" height="20" rx="4"/></svg>
        </div>
        <h3 class="card-title">Business / ISP Kit</h3>
        <p class="card-desc">Launch and operate your own ISP business end-to-end.</p>
        <div class="card-footer">
          <span class="card-price">$149</span>
          <span class="card-action">View →</span>
        </div>
      </a>
    </div>

  </div>
</section>

<style>
/* Cards grid */
.cards-grid {
  display: grid;
  gap: 1rem;
}
@media (min-width: 768px) {
  .cards-grid { grid-template-columns: repeat(3, 1fr); }
}

/* Card */
.card {
  position: relative;
  display: flex;
  flex-direction: column;
  padding: 1.5rem;         /* p-6 */
  border-radius: 1rem;     /* rounded-2xl */
  border: 1px solid var(--color-border);
  background: var(--color-bg);
  text-decoration: none;
  color: inherit;
  transition: all 0.2s ease;
}
.card:hover {
  box-shadow: 0 10px 15px -3px rgba(0,0,0,0.1), 0 4px 6px -4px rgba(0,0,0,0.1);
  transform: translateY(-0.25rem);  /* -translate-y-1 */
}

/* Card badge */
.card-badge {
  position: absolute;
  top: 1rem;
  right: 1rem;
}

/* Card icon box */
.card-icon-box {
  width: 2.5rem;          /* 40px */
  height: 2.5rem;
  border-radius: 0.75rem;  /* rounded-xl */
  background: var(--color-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
  color: var(--color-muted-fg);
}

/* Card content */
.card-title {
  font-family: var(--font-heading);
  font-size: 1rem;         /* text-base */
  font-weight: 700;
  margin-bottom: 0.375rem;
}
.card-desc {
  font-size: 0.875rem;     /* text-sm */
  color: var(--color-muted-fg);
  line-height: 1.625;
  margin-bottom: 1.25rem;
  flex: 1;
}

/* Card footer */
.card-footer {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border);
}
.card-price {
  font-family: var(--font-heading);
  font-size: 1.25rem;      /* text-xl */
  font-weight: 700;
}
.card-action {
  font-size: 0.875rem;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}
</style>
```

**Rules:**
- 3 cards in a row at 768px+, stacks on mobile
- Cards are `<a>` links (whole card clickable) — `text-decoration: none; color: inherit`
- Hover: lifts up 0.25rem + shadow
- Icon box: `2.5rem` square, gray background (`var(--color-secondary)`)
- Badge: absolute positioned `top:1rem; right:1rem`
- Footer: flex row with top border separating price from "View →"

---

## 7. SECTION RECIPE: CTA Banner (Gray Background)

```html
<!-- CTA BANNER SECTION -->
<section style="background: var(--color-secondary);" class="section">
  <div class="container">

    <div class="cta-center">
      <h2 class="cta-headline">Start learning for free</h2>
      <p class="cta-desc">
        Master the fundamentals, then upgrade to full solutions when you're ready.
      </p>
      <div class="cta-buttons">
        <a href="https://example.com" target="_blank" rel="noopener noreferrer" class="btn btn-outline btn-icon">
          Visit external site <span style="font-size:1rem;">↗</span>
        </a>
        <a href="/products" class="btn btn-primary">Upgrade to Full Solutions</a>
      </div>
    </div>

  </div>
</section>

<style>
.cta-center {
  max-width: 42rem;          /* max-w-2xl */
  margin: 0 auto;
  text-align: center;
}
.cta-headline {
  font-family: var(--font-heading);
  font-size: 1.5rem;         /* text-2xl */
  font-weight: 700;
  letter-spacing: -0.025em;
  margin-bottom: 0.75rem;
}
@media (min-width: 768px) {
  .cta-headline { font-size: 1.875rem; }   /* md:text-3xl */
}
.cta-desc {
  color: var(--color-muted-fg);
  margin-bottom: 2rem;
  font-size: 0.875rem;
}
.cta-buttons {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  justify-content: center;
}
@media (min-width: 640px) {
  .cta-buttons { flex-direction: row; }
}
</style>
```

**Rules:**
- Background: `var(--color-secondary)` (96% gray)
- Content centered: `max-width: 42rem; margin: 0 auto; text-align: center`
- Buttons stack vertically on mobile, side-by-side at 640px+
- Primary: filled dark button, Secondary: outlined white button

---

## 8. ANIMATIONS (Pure CSS)

All animations use only CSS — no JavaScript library needed.

### Entrance from below (fade + slide up)
```css
@keyframes fade-up {
  from { opacity: 0; transform: translateY(20px); }
  to   { opacity: 1; transform: translateY(0); }
}
.animate-fade-up {
  animation: fade-up 0.6s ease-out both;
}
```

### Entrance with scale (images)
```css
@keyframes fade-scale {
  from { opacity: 0; transform: scale(0.92); }
  to   { opacity: 1; transform: scale(1); }
}
.animate-fade-scale {
  animation: fade-scale 0.7s ease-out 0.2s both;
}
```

### Staggered entrance (cards — add via inline delays)
```css
@keyframes fade-in-up {
  from { opacity: 0; transform: translateY(20px); }
  to   { opacity: 1; transform: translateY(0); }
}
.animate-fade-in-up {
  opacity: 0;
  animation: fade-in-up 0.4s ease-out both;
}
```
```html
<div class="card animate-fade-in-up" style="animation-delay: 0.00s;">...</div>
<div class="card animate-fade-in-up" style="animation-delay: 0.10s;">...</div>
<div class="card animate-fade-in-up" style="animation-delay: 0.20s;">...</div>
```

### Floating bounce (infinite loop — for badges/decorations)
```css
@keyframes float-bounce {
  0%, 100% { transform: translateY(0); }
  50%      { transform: translateY(-8px); }
}
.animate-float {
  animation: float-bounce 3s ease-in-out infinite;
}
```

### Scroll-triggered animations (vanilla JS Intersection Observer)
```js
// Add to your script: reveals elements when they scroll into view
const observer = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    if (entry.isIntersecting) {
      entry.target.classList.add('revealed');
    } // remove 'if' to animate once; remove the closing brace to repeat
  });
}, { threshold: 0.1 });

document.querySelectorAll('.reveal').forEach(el => observer.observe(el));
```
```css
.reveal {
  opacity: 0;
  transform: translateY(20px);
  transition: opacity 0.6s ease-out, transform 0.6s ease-out;
}
.reveal.revealed {
  opacity: 1;
  transform: translateY(0);
}
```
```html
<div class="reveal" style="transition-delay: 0.1s;">This fades in on scroll</div>
```

---

## 9. TYPOGRAPHY HIERARCHY

| Element | CSS |
|---|---|
| Page headline | `font-family:'Poppins',sans-serif; font-size:1.875rem; font-weight:700; line-height:1.1; letter-spacing:-0.025em;` |
| (md+) | `font-size:3rem` |
| (lg+) | `font-size:3.5rem` |
| Section title | `font-family:'Poppins',sans-serif; font-size:1.5rem; font-weight:700; letter-spacing:-0.025em;` |
| (md+) | `font-size:1.875rem` |
| Section subtitle | `font-size:0.875rem; color:#737373;` |
| Card title | `font-family:'Poppins',sans-serif; font-size:1rem; font-weight:700;` |
| Card description | `font-size:0.875rem; color:#737373; line-height:1.625;` |
| Body paragraph | `font-size:1rem; color:#737373;` |
| Small label | `font-size:0.75rem; font-weight:500;` |

---

## 10. COLOR USAGE

| CSS Variable | Hex Equivalent | Use |
|---|---|---|
| `var(--color-bg)` | `#ffffff` | Page background, cards, stat boxes |
| `var(--color-fg)` | `#141414` | Headlines, labels, link text |
| `var(--color-muted-fg)` | `#737373` | All secondary text: descriptions, subtitles |
| `var(--color-elevated)` | `#f7f7f7` | Hero section background |
| `var(--color-secondary)` | `#f5f5f5` | CTA banner background, icon containers, badges |
| `var(--color-border)` | `#e5e5e5` | Card borders, image borders, footer border |
| `var(--color-brand-red)` | `hsl(0,80%,50%)` | Accent red |
| `var(--color-brand-blue)` | `hsl(217,91%,55%)` | Accent blue |

---

## 11. SPACING SCALE

| CSS Value | Where Used |
|---|---|
| `gap: 1rem` | Between grid items, card layouts |
| `margin-bottom: 1rem` | Below headlines |
| `margin-bottom: 2rem` | Below paragraphs before CTAs |
| `padding: 3rem 0` / `5rem 0` (md+) | Standard section vertical padding |
| `padding: 2.5rem 0` / `5rem 0` / `7rem 0` (lg+) | Hero section (larger) |
| `padding: 0 1.5rem` | Container horizontal padding |
| `padding: 1.5rem` | Card inner padding |
| `padding-top: 1rem` | Separator above card footer |
| `border-radius: 1rem` | Cards & stat boxes |
| `border-radius: 0.75rem` | Icon containers |
| `border-radius: 9999px` | Buttons, CTAs, category images |

---

## 12. QUICK-START: Single HTML File Template

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>My Landing Page</title>
  <link rel="icon" href="/favicon.png" />
  <style>
    /* ===== PASTE SECTION 2 (Global Styles) HERE ===== */

    /* ===== PASTE ANIMATION KEYFRAMES (Section 8) HERE ===== */

    /* ===== PASTE SECTION CSS (4, 5, 6, 7) HERE ===== */
  </style>
</head>
<body>

  <!-- ===== PASTE SECTION 1 (Splash Screen) HERE ===== -->

  <!-- ===== PASTE SECTION 3 (Page Layout) HERE ===== -->

  <script>
    // Intersection Observer for scroll-triggered animations (Section 8)
    const observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          entry.target.classList.add('revealed');
        }
      });
    }, { threshold: 0.1 });
    document.querySelectorAll('.reveal').forEach(el => observer.observe(el));
  </script>

</body>
</html>
```

**That's it.** One HTML file, one CSS block at the top, zero frameworks, zero build tools. All sections are standalone — copy only the parts you need.
