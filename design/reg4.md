<!DOCTYPE html>

<html class="light" lang="en"><head>
<meta charset="utf-8"/>
<meta content="width=device-width, initial-scale=1.0" name="viewport"/>
<title>TrustCore Portal - Registration Submitted</title>
<script src="https://cdn.tailwindcss.com?plugins=forms,container-queries"></script>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700&amp;family=Public+Sans:wght@500;600;700;800&amp;display=swap" rel="stylesheet"/>
<link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&amp;display=swap" rel="stylesheet"/>
<link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&amp;display=swap" rel="stylesheet"/>
<script id="tailwind-config">
      tailwind.config = {
        darkMode: "class",
        theme: {
          extend: {
            "colors": {
                    "on-secondary": "#ffffff",
                    "secondary": "#515f74",
                    "secondary-fixed": "#d5e3fd",
                    "on-secondary-fixed": "#0d1c2f",
                    "error": "#ba1a1a",
                    "surface-container": "#eceef0",
                    "on-primary": "#ffffff",
                    "inverse-primary": "#bec6e0",
                    "on-secondary-container": "#57657b",
                    "surface": "#f7f9fb",
                    "on-primary-fixed": "#131b2e",
                    "secondary-container": "#d5e3fd",
                    "surface-tint": "#565e74",
                    "tertiary-fixed": "#6ffbbe",
                    "surface-container-low": "#f2f4f6",
                    "surface-container-highest": "#e0e3e5",
                    "on-tertiary-fixed": "#002113",
                    "tertiary": "#000000",
                    "on-surface-variant": "#45464d",
                    "surface-bright": "#f7f9fb",
                    "on-primary-fixed-variant": "#3f465c",
                    "on-primary-container": "#7c839b",
                    "background": "#f7f9fb",
                    "surface-dim": "#d8dadc",
                    "primary": "#000000",
                    "on-surface": "#191c1e",
                    "primary-container": "#131b2e",
                    "tertiary-container": "#002113",
                    "primary-fixed-dim": "#bec6e0",
                    "inverse-surface": "#2d3133",
                    "tertiary-fixed-dim": "#4edea3",
                    "surface-container-lowest": "#ffffff",
                    "on-tertiary-container": "#009668",
                    "outline": "#76777d",
                    "inverse-on-surface": "#eff1f3",
                    "primary-fixed": "#dae2fd",
                    "error-container": "#ffdad6",
                    "secondary-fixed-dim": "#b9c7e0",
                    "on-tertiary-fixed-variant": "#005236",
                    "surface-container-high": "#e6e8ea",
                    "outline-variant": "#c6c6cd",
                    "on-error-container": "#93000a",
                    "on-tertiary": "#ffffff",
                    "on-background": "#191c1e",
                    "on-secondary-fixed-variant": "#3a485c",
                    "surface-variant": "#e0e3e5",
                    "on-error": "#ffffff"
            },
            "borderRadius": {
                    "DEFAULT": "0.125rem",
                    "lg": "0.25rem",
                    "xl": "0.5rem",
                    "full": "0.75rem"
            },
            "spacing": {
                    "base": "8px",
                    "gutter": "24px",
                    "margin": "32px",
                    "container-max": "1120px",
                    "form-gap": "20px"
            },
            "fontFamily": {
                    "headline-md": ["Public Sans"],
                    "headline-lg": ["Public Sans"],
                    "body-sm": ["Inter"],
                    "body-md": ["Inter"],
                    "label-bold": ["Inter"],
                    "label-caps": ["Inter"]
            },
            "fontSize": {
                    "headline-md": ["24px", {"lineHeight": "32px", "letterSpacing": "-0.01em", "fontWeight": "600"}],
                    "headline-lg": ["30px", {"lineHeight": "38px", "letterSpacing": "-0.02em", "fontWeight": "700"}],
                    "body-sm": ["14px", {"lineHeight": "20px", "fontWeight": "400"}],
                    "body-md": ["16px", {"lineHeight": "24px", "fontWeight": "400"}],
                    "label-bold": ["14px", {"lineHeight": "16px", "letterSpacing": "0.01em", "fontWeight": "600"}],
                    "label-caps": ["12px", {"lineHeight": "16px", "letterSpacing": "0.05em", "fontWeight": "700"}]
            }
          },
        },
      }
    </script>
<style>
        .material-symbols-outlined {
            font-variation-settings: 'FILL' 0, 'wght' 400, 'GRAD' 0, 'opsz' 24;
        }
        body {
            background-color: #f7f9fb;
        }
    </style>
</head>
<body class="bg-background text-on-background antialiased">
<!-- TopAppBar -->
<header class="bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 shadow-sm docked full-width top-0 z-50">
<div class="flex items-center justify-between w-full max-w-[1120px] mx-auto h-16 px-6">
<div class="text-xl font-bold tracking-tight text-slate-900 dark:text-white font-headline-lg">
                TrustCore Portal
            </div>
<nav class="hidden md:flex space-x-8">
<a class="text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 font-public-sans text-sm font-medium tracking-tight hover:bg-slate-50 dark:hover:bg-slate-800 transition-all duration-200 px-3 py-2 rounded" href="#">Dashboard</a>
<a class="text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 font-public-sans text-sm font-medium tracking-tight hover:bg-slate-50 dark:hover:bg-slate-800 transition-all duration-200 px-3 py-2 rounded" href="#">Compliance</a>
<a class="text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 font-public-sans text-sm font-medium tracking-tight hover:bg-slate-50 dark:hover:bg-slate-800 transition-all duration-200 px-3 py-2 rounded" href="#">Documents</a>
</nav>
<div class="flex items-center gap-4">
<button class="text-slate-600 dark:text-slate-400 font-public-sans text-sm font-medium tracking-tight hover:bg-slate-50 dark:hover:bg-slate-800 transition-all duration-200 px-3 py-2 rounded">Help &amp; Support</button>
</div>
</div>
</header>
<main class="max-w-[1120px] mx-auto px-6 py-16 md:py-24">
<!-- Success Container -->
<div class="max-w-2xl mx-auto bg-white border border-slate-200 rounded shadow-[0px_1px_3px_rgba(15,23,42,0.1)] p-8 md:p-12 text-center">
<!-- Stepper Progress (Completed View) -->
<div class="flex items-center justify-center mb-12">
<div class="flex items-center">
<div class="w-8 h-8 rounded-full bg-on-tertiary-container text-white flex items-center justify-center shadow-sm">
<span class="material-symbols-outlined text-lg">check</span>
</div>
<div class="w-16 h-1 bg-on-tertiary-container"></div>
</div>
<div class="flex items-center">
<div class="w-8 h-8 rounded-full bg-on-tertiary-container text-white flex items-center justify-center shadow-sm">
<span class="material-symbols-outlined text-lg">check</span>
</div>
<div class="w-16 h-1 bg-on-tertiary-container"></div>
</div>
<div class="flex items-center">
<div class="w-8 h-8 rounded-full bg-on-tertiary-container text-white flex items-center justify-center shadow-sm">
<span class="material-symbols-outlined text-lg">check</span>
</div>
<div class="w-16 h-1 bg-on-tertiary-container"></div>
</div>
<div class="w-8 h-8 rounded-full bg-on-tertiary-container text-white flex items-center justify-center shadow-sm">
<span class="material-symbols-outlined text-lg">check</span>
</div>
</div>
<!-- Illustration/Visual Anchor -->
<div class="mb-8 flex justify-center">
<div class="relative">
<div class="w-24 h-24 bg-surface-container-low rounded-full flex items-center justify-center">
<span class="material-symbols-outlined text-5xl text-on-tertiary-container" style="font-variation-settings: 'FILL' 1;">verified_user</span>
</div>
<div class="absolute -bottom-1 -right-1 bg-white p-1 rounded-full border border-slate-200">
<span class="material-symbols-outlined text-on-tertiary-container">check_circle</span>
</div>
</div>
</div>
<!-- Content -->
<h1 class="font-headline-lg text-headline-lg text-slate-900 mb-4">Registration Submitted</h1>
<div class="inline-flex items-center px-3 py-1 bg-secondary-fixed text-on-secondary-fixed rounded-full font-label-caps text-label-caps mb-6">
<span class="material-symbols-outlined text-sm mr-1.5" style="font-variation-settings: 'FILL' 1;">pending</span>
                Pending Verification
            </div>
<p class="font-body-md text-body-md text-slate-600 mb-10 max-w-lg mx-auto">
                Thank you for registering. Your details and documents are currently being reviewed by our team. You will receive an email notification once the verification process is complete.
            </p>
<!-- Action Buttons -->
<div class="flex flex-col gap-4 max-w-sm mx-auto">
<button class="w-full bg-primary-container text-white py-3.5 px-6 rounded font-label-bold text-label-bold hover:opacity-90 transition-opacity flex items-center justify-center gap-2">
                    View Registration Status
                    <span class="material-symbols-outlined text-lg">arrow_forward</span>
</button>
<a class="w-full border border-slate-300 text-slate-700 py-3.5 px-6 rounded font-label-bold text-label-bold hover:bg-slate-50 transition-colors flex items-center justify-center" href="#">
                    Contact Support
                </a>
</div>
<!-- Info Box -->
<div class="mt-12 p-6 bg-surface-container-lowest border border-slate-100 rounded text-left">
<div class="flex gap-4">
<span class="material-symbols-outlined text-on-secondary-container">info</span>
<div>
<h4 class="font-label-bold text-label-bold text-slate-900 mb-1">What's next?</h4>
<p class="font-body-sm text-body-sm text-slate-500">Typical review time is 24-48 business hours. You may be contacted if additional documentation is required for your specific jurisdiction.</p>
</div>
</div>
</div>
</div>
<!-- Secondary Content: Reassurance Grid -->
<div class="mt-16 grid grid-cols-1 md:grid-cols-3 gap-8">
<div class="p-6 bg-white border border-slate-200 rounded">
<span class="material-symbols-outlined text-slate-400 mb-4 text-3xl">security</span>
<h3 class="font-label-bold text-label-bold text-slate-900 mb-2">Secure Storage</h3>
<p class="font-body-sm text-body-sm text-slate-500">All uploaded documents are encrypted with AES-256 standards in our institutional-grade vault.</p>
</div>
<div class="p-6 bg-white border border-slate-200 rounded">
<span class="material-symbols-outlined text-slate-400 mb-4 text-3xl">history_edu</span>
<h3 class="font-label-bold text-label-bold text-slate-900 mb-2">Audit Ready</h3>
<p class="font-body-sm text-body-sm text-slate-500">Every step of your onboarding journey is logged for compliance and regulatory reporting requirements.</p>
</div>
<div class="p-6 bg-white border border-slate-200 rounded">
<span class="material-symbols-outlined text-slate-400 mb-4 text-3xl">support_agent</span>
<h3 class="font-label-bold text-label-bold text-slate-900 mb-2">Priority Review</h3>
<p class="font-body-sm text-body-sm text-slate-500">As a high-stakes B2B partner, your application is placed in our priority verification queue.</p>
</div>
</div>
<!-- Decorative Background Element -->
<div class="fixed inset-0 -z-10 pointer-events-none opacity-40">
<div class="absolute top-0 right-0 w-[500px] h-[500px] bg-secondary-fixed blur-[120px] rounded-full transform translate-x-1/2 -translate-y-1/2"></div>
<div class="absolute bottom-0 left-0 w-[400px] h-[400px] bg-on-tertiary-container/10 blur-[100px] rounded-full transform -translate-x-1/2 translate-y-1/2"></div>
</div>
</main>
<footer class="w-full max-w-[1120px] mx-auto px-6 py-8 border-t border-slate-200 mt-auto">
<div class="flex flex-col md:flex-row justify-between items-center gap-4">
<p class="font-body-sm text-body-sm text-slate-500">© 2024 TrustCore Institutional Services. All rights reserved.</p>
<div class="flex gap-6">
<a class="font-body-sm text-body-sm text-slate-500 hover:text-slate-900" href="#">Privacy Policy</a>
<a class="font-body-sm text-body-sm text-slate-500 hover:text-slate-900" href="#">Terms of Service</a>
</div>
</div>
</footer>
<!-- Abstract Image Description (Hidden) -->
<div class="hidden">
<img data-alt="A clean, minimalist digital environment representing institutional stability and security. The scene is bathed in soft, high-key lighting with a palette of crisp whites, slate greys, and corporate blues. Subtle geometric patterns suggest a structured network or grid, creating a sense of order and technological precision. The overall mood is calm, professional, and reassuring, perfectly aligning with a successful B2B compliance submission." src="https://lh3.googleusercontent.com/aida-public/AB6AXuAmrS7QT-VPuWYpvKHYxk2FJuvJVs3Gj36NYvBtvDVy430fTUX6q23DhjANcyzxsZODRtGhvEwVIltbWl1lxcl4F1NFY3ySfHxPZ-h1fs9HfXbnguDQCkLua_-7wB-U3iHaR_cbEirEt3ZiALRp9ap47oUE8u62CGErHKi7Z_J4IrEAb2CQnJ5KZcSufmuAIp1ww4Sw5m71vHphPr7t8Vg9c_0fBeZd1ZwkrEJo86Dxk57UjVxaaOoSph5ec5lAaWOyBX1EbBNOMNo"/>
</div>
</body></html>