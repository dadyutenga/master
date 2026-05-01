<!DOCTYPE html>

<html class="light" lang="en"><head>
<meta charset="utf-8"/>
<meta content="width=device-width, initial-scale=1.0" name="viewport"/>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700&amp;family=Public+Sans:wght@500;600;700;800&amp;display=swap" rel="stylesheet"/>
<link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&amp;display=swap" rel="stylesheet"/>
<link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&amp;display=swap" rel="stylesheet"/>
<script src="https://cdn.tailwindcss.com?plugins=forms,container-queries"></script>
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
<body class="font-body-md text-on-background">
<!-- TopAppBar from JSON -->
<header class="bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 shadow-sm docked full-width top-0 z-50">
<div class="flex items-center justify-between w-full max-w-[1120px] mx-auto h-16 px-6">
<div class="text-xl font-bold tracking-tight text-slate-900 dark:text-white font-headline-md">
                TrustCore Portal
            </div>
<nav class="hidden md:flex items-center space-x-8 h-full">
<a class="text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 font-public-sans text-sm font-medium tracking-tight transition-all duration-200 h-full flex items-center px-1" href="#">Overview</a>
<a class="text-blue-700 dark:text-blue-400 font-semibold border-b-2 border-blue-700 font-public-sans text-sm tracking-tight transition-all duration-200 h-full flex items-center px-1" href="#">Registration</a>
<a class="text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 font-public-sans text-sm font-medium tracking-tight transition-all duration-200 h-full flex items-center px-1" href="#">Compliance</a>
<a class="text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 font-public-sans text-sm font-medium tracking-tight transition-all duration-200 h-full flex items-center px-1" href="#">Documents</a>
</nav>
<div class="flex items-center space-x-4">
<button class="text-slate-600 dark:text-slate-400 font-public-sans text-sm font-medium hover:bg-slate-50 dark:hover:bg-slate-800 px-3 py-2 rounded-lg transition-all duration-200">
                    Help &amp; Support
                </button>
<div class="h-8 w-8 rounded-full bg-secondary-container flex items-center justify-center text-on-secondary-fixed font-bold text-xs">
                    JD
                </div>
</div>
</div>
</header>
<main class="max-w-[1120px] mx-auto px-6 py-12">
<!-- Progress Stepper -->
<div class="mb-12">
<div class="flex items-center justify-between max-w-3xl mx-auto">
<div class="flex flex-col items-center flex-1 relative">
<div class="w-10 h-10 rounded-full bg-on-tertiary-container text-white flex items-center justify-center z-10">
<span class="material-symbols-outlined text-base">check</span>
</div>
<span class="mt-2 font-label-bold text-on-surface">Entity Info</span>
<div class="absolute top-5 left-1/2 w-full h-1 bg-on-tertiary-container -z-0"></div>
</div>
<div class="flex flex-col items-center flex-1 relative">
<div class="w-10 h-10 rounded-full bg-blue-700 text-white flex items-center justify-center z-10 border-4 border-blue-100">
<span class="font-bold">2</span>
</div>
<span class="mt-2 font-label-bold text-blue-700">Hotel</span>
<div class="absolute top-5 left-1/2 w-full h-1 bg-slate-200 -z-0"></div>
</div>
<div class="flex flex-col items-center flex-1 relative">
<div class="w-10 h-10 rounded-full bg-white border-2 border-slate-300 text-slate-400 flex items-center justify-center z-10">
<span class="font-bold">3</span>
</div>
<span class="mt-2 font-label-bold text-slate-500">Contact</span>
<div class="absolute top-5 left-1/2 w-full h-1 bg-slate-200 -z-0"></div>
</div>
<div class="flex flex-col items-center flex-1">
<div class="w-10 h-10 rounded-full bg-white border-2 border-slate-300 text-slate-400 flex items-center justify-center">
<span class="font-bold">4</span>
</div>
<span class="mt-2 font-label-bold text-slate-500">Review</span>
</div>
</div>
</div>
<!-- Form Container -->
<div class="max-w-3xl mx-auto">
<div class="bg-white border border-slate-200 p-10 rounded-lg shadow-[0px_1px_3px_rgba(15,23,42,0.1)]">
<div class="mb-8">
<h1 class="font-headline-lg text-headline-lg text-on-surface mb-2">Hotel Information</h1>
<p class="font-body-md text-secondary">Please provide the specific details of the hospitality property being registered under this account.</p>
</div>
<form class="space-y-[20px]">
<!-- Row 1: Name & Category -->
<div class="grid grid-cols-1 md:grid-cols-2 gap-[20px]">
<div class="flex flex-col space-y-2">
<label class="font-label-bold text-on-surface-variant" for="hotel-name">Hotel Name</label>
<input class="border border-slate-300 rounded-lg p-3 focus:border-blue-700 focus:ring-1 focus:ring-blue-700 outline-none transition-all placeholder:text-slate-400" id="hotel-name" placeholder="Enter legal property name" type="text"/>
</div>
<div class="flex flex-col space-y-2">
<label class="font-label-bold text-on-surface-variant" for="category">Category</label>
<select class="border border-slate-300 rounded-lg p-3 focus:border-blue-700 focus:ring-1 focus:ring-blue-700 outline-none transition-all bg-white text-on-surface" id="category">
<option disabled="" selected="" value="">Select property type</option>
<option value="luxury">Luxury</option>
<option value="boutique">Boutique</option>
<option value="business">Business</option>
<option value="budget">Budget</option>
<option value="resort">Resort</option>
</select>
</div>
</div>
<!-- Row 2: Room Count -->
<div class="grid grid-cols-1 md:grid-cols-3 gap-[20px]">
<div class="flex flex-col space-y-2">
<label class="font-label-bold text-on-surface-variant" for="room-count">Room Count</label>
<div class="relative">
<input class="w-full border border-slate-300 rounded-lg p-3 focus:border-blue-700 focus:ring-1 focus:ring-blue-700 outline-none transition-all" id="room-count" placeholder="0" type="number"/>
<span class="absolute right-3 top-3 text-slate-400 material-symbols-outlined text-base">bed</span>
</div>
</div>
</div>
<!-- Row 3: Physical Address -->
<div class="flex flex-col space-y-2">
<label class="font-label-bold text-on-surface-variant" for="address">Physical Address</label>
<input class="border border-slate-300 rounded-lg p-3 focus:border-blue-700 focus:ring-1 focus:ring-blue-700 outline-none transition-all" id="address" placeholder="Street address, suite, or floor" type="text"/>
</div>
<!-- Row 4: City & Country -->
<div class="grid grid-cols-1 md:grid-cols-2 gap-[20px]">
<div class="flex flex-col space-y-2">
<label class="font-label-bold text-on-surface-variant" for="city">City</label>
<input class="border border-slate-300 rounded-lg p-3 focus:border-blue-700 focus:ring-1 focus:ring-blue-700 outline-none transition-all" id="city" placeholder="City" type="text"/>
</div>
<div class="flex flex-col space-y-2">
<label class="font-label-bold text-on-surface-variant" for="country">Country</label>
<select class="border border-slate-300 rounded-lg p-3 focus:border-blue-700 focus:ring-1 focus:ring-blue-700 outline-none transition-all bg-white" id="country">
<option disabled="" selected="" value="">Select country</option>
<option value="us">United States</option>
<option value="uk">United Kingdom</option>
<option value="ca">Canada</option>
<option value="fr">France</option>
<option value="jp">Japan</option>
</select>
</div>
</div>
<!-- Action Area -->
<div class="flex items-center justify-between pt-8 mt-4 border-t border-slate-100">
<button class="flex items-center space-x-2 px-6 py-3 border border-slate-300 rounded-lg font-label-bold text-secondary hover:bg-slate-50 active:opacity-80 transition-all" type="button">
<span class="material-symbols-outlined text-base">arrow_back</span>
<span>Back</span>
</button>
<button class="flex items-center space-x-2 px-8 py-3 bg-blue-800 text-white rounded-lg font-label-bold hover:bg-blue-900 active:opacity-80 shadow-md transition-all" type="submit">
<span>Next Step</span>
<span class="material-symbols-outlined text-base">arrow_forward</span>
</button>
</div>
</form>
</div>
<!-- Contextual Help Card -->
<div class="mt-8 grid grid-cols-1 md:grid-cols-2 gap-6">
<div class="flex items-start p-4 bg-slate-50 border border-slate-200 rounded-lg">
<div class="p-2 bg-white rounded-md shadow-sm mr-4">
<span class="material-symbols-outlined text-blue-700">verified_user</span>
</div>
<div>
<h4 class="font-label-bold text-on-surface mb-1">Security Verified</h4>
<p class="text-xs text-secondary leading-relaxed">Your data is encrypted using institutional-grade protocols during the onboarding process.</p>
</div>
</div>
<div class="flex items-start p-4 bg-slate-50 border border-slate-200 rounded-lg">
<div class="p-2 bg-white rounded-md shadow-sm mr-4">
<span class="material-symbols-outlined text-blue-700">help</span>
</div>
<div>
<h4 class="font-label-bold text-on-surface mb-1">Need Assistance?</h4>
<p class="text-xs text-secondary leading-relaxed">Contact our B2B support team for expedited verification or large portfolio uploads.</p>
</div>
</div>
</div>
</div>
</main>
<!-- Footer-like element -->
<footer class="max-w-[1120px] mx-auto px-6 py-12 flex flex-col md:flex-row items-center justify-between border-t border-slate-200 mt-12 opacity-60">
<div class="flex items-center space-x-2 mb-4 md:mb-0">
<span class="font-headline-md text-sm font-bold">TrustCore</span>
<span class="text-xs font-body-sm">© 2024 TrustCore Institutional Compliance</span>
</div>
<div class="flex space-x-6 text-xs font-body-sm">
<a class="hover:underline" href="#">Privacy Policy</a>
<a class="hover:underline" href="#">Terms of Service</a>
<a class="hover:underline" href="#">System Status</a>
</div>
</footer>
</body></html>