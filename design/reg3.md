<!DOCTYPE html>

<html class="light" lang="en"><head>
<meta charset="utf-8"/>
<meta content="width=device-width, initial-scale=1.0" name="viewport"/>
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
                }
            }
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
<body class="font-body-md text-on-background selection:bg-primary-fixed selection:text-on-primary-fixed">
<header class="bg-white dark:bg-slate-900 shadow-sm border-b border-slate-200 dark:border-slate-800 sticky top-0 z-50">
<div class="flex items-center justify-between w-full max-w-[1120px] mx-auto h-16 px-6">
<div class="text-xl font-bold tracking-tight text-slate-900 dark:text-white font-public-sans">
                TrustCore Portal
            </div>
<nav class="hidden md:flex items-center gap-8 h-full">
<a class="flex items-center h-full px-1 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 font-public-sans text-sm font-medium tracking-tight transition-all duration-200" href="#">Company Info</a>
<a class="flex items-center h-full px-1 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 font-public-sans text-sm font-medium tracking-tight transition-all duration-200" href="#">Legal Structure</a>
<a class="flex items-center h-full px-1 text-blue-700 dark:text-blue-400 font-semibold border-b-2 border-blue-700 font-public-sans text-sm tracking-tight" href="#">Documents</a>
<a class="flex items-center h-full px-1 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 font-public-sans text-sm font-medium tracking-tight transition-all duration-200" href="#">Help &amp; Support</a>
</nav>
<div class="flex items-center gap-4">
<span class="material-symbols-outlined text-secondary cursor-pointer" data-icon="account_circle">account_circle</span>
</div>
</div>
</header>
<main class="max-w-[1120px] mx-auto px-6 py-12">
<div class="mb-12">
<div class="flex items-center justify-between mb-8 overflow-x-auto pb-4">
<div class="flex items-center min-w-max">
<div class="flex items-center text-on-tertiary-container">
<div class="w-8 h-8 rounded-full bg-on-tertiary-container flex items-center justify-center text-white font-bold text-sm">
<span class="material-symbols-outlined text-[18px]" data-icon="check" data-weight="fill" style="font-variation-settings: 'FILL' 1;">check</span>
</div>
<span class="ml-3 font-label-bold text-on-surface">Company Profile</span>
</div>
<div class="w-16 h-[2px] bg-on-tertiary-container mx-4"></div>
<div class="flex items-center text-on-tertiary-container">
<div class="w-8 h-8 rounded-full bg-on-tertiary-container flex items-center justify-center text-white font-bold text-sm">
<span class="material-symbols-outlined text-[18px]" data-icon="check" data-weight="fill" style="font-variation-settings: 'FILL' 1;">check</span>
</div>
<span class="ml-3 font-label-bold text-on-surface">Compliance</span>
</div>
<div class="w-16 h-[2px] bg-on-tertiary-container mx-4"></div>
<div class="flex items-center text-primary">
<div class="w-8 h-8 rounded-full border-2 border-primary flex items-center justify-center bg-primary text-white font-bold text-sm">3</div>
<span class="ml-3 font-label-bold text-primary">Documents</span>
</div>
<div class="w-16 h-[2px] bg-outline-variant mx-4"></div>
<div class="flex items-center text-outline">
<div class="w-8 h-8 rounded-full border-2 border-outline-variant flex items-center justify-center text-outline font-bold text-sm">4</div>
<span class="ml-3 font-label-bold text-outline">Review</span>
</div>
</div>
</div>
<div class="mb-4">
<h1 class="font-headline-lg text-headline-lg text-on-surface mb-2">Contact &amp; Document Upload</h1>
<p class="font-body-md text-secondary max-w-2xl">Please provide the primary contact details for your organization and upload the required regulatory certificates to complete your institutional verification.</p>
</div>
<div class="w-full bg-surface-container h-1.5 rounded-full overflow-hidden mb-12">
<div class="bg-on-tertiary-container h-full w-3/4"></div>
</div>
</div>
<div class="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
<section class="lg:col-span-7 space-y-8">
<div class="bg-white border border-outline-variant p-8 rounded-lg shadow-sm">
<h2 class="font-headline-md text-headline-md text-on-surface mb-6 flex items-center gap-2">
<span class="material-symbols-outlined" data-icon="contact_mail">contact_mail</span>
                        Primary Contact Details
                    </h2>
<div class="space-y-5">
<div class="flex flex-col gap-2">
<label class="font-label-bold text-on-surface" for="full_name">Primary Contact Full Name</label>
<div class="relative group">
<input class="w-full h-12 px-4 rounded border border-outline hover:border-on-surface focus:border-blue-700 focus:ring-1 focus:ring-blue-700 outline-none transition-all duration-200" id="full_name" placeholder="e.g. John Doe" type="text"/>
<div class="hidden absolute right-4 top-1/2 -translate-y-1/2 text-on-tertiary-container">
<span class="material-symbols-outlined text-[20px]" data-icon="check_circle" data-weight="fill" style="font-variation-settings: 'FILL' 1;">check_circle</span>
</div>
</div>
</div>
<div class="grid grid-cols-1 md:grid-cols-2 gap-5">
<div class="flex flex-col gap-2">
<label class="font-label-bold text-on-surface" for="phone">Phone Number</label>
<div class="relative">
<span class="absolute left-4 top-1/2 -translate-y-1/2 text-outline font-label-bold">+255</span>
<input class="w-full h-12 pl-14 pr-4 rounded border border-outline hover:border-on-surface focus:border-blue-700 focus:ring-1 focus:ring-blue-700 outline-none transition-all duration-200" id="phone" placeholder="712 345 678" type="tel"/>
</div>
</div>
<div class="flex flex-col gap-2">
<label class="font-label-bold text-on-surface" for="email">Email Address</label>
<input class="w-full h-12 px-4 rounded border border-outline hover:border-on-surface focus:border-blue-700 focus:ring-1 focus:ring-blue-700 outline-none transition-all duration-200" id="email" placeholder="contact@company.co.tz" type="email"/>
</div>
</div>
</div>
</div>
<div class="bg-white border border-outline-variant p-8 rounded-lg shadow-sm">
<h2 class="font-headline-md text-headline-md text-on-surface mb-6 flex items-center gap-2">
<span class="material-symbols-outlined" data-icon="upload_file">upload_file</span>
                        Required Certificates
                    </h2>
<div class="space-y-6">
<div class="flex flex-col gap-3">
<div class="flex items-center justify-between">
<span class="font-label-bold text-on-surface">BRELA Certificate</span>
<span class="font-label-caps text-on-secondary-container bg-secondary-container px-2 py-0.5 rounded">MANDATORY</span>
</div>
<div class="border-2 border-dashed border-outline-variant rounded-lg p-10 bg-surface hover:bg-secondary-container/10 hover:border-blue-700 transition-all cursor-pointer group flex flex-col items-center justify-center gap-4">
<div class="w-12 h-12 rounded-full bg-surface-container flex items-center justify-center group-hover:scale-110 transition-transform">
<span class="material-symbols-outlined text-outline group-hover:text-blue-700" data-icon="image">image</span>
</div>
<div class="text-center">
<p class="font-label-bold text-on-surface">Click to upload or drag and drop</p>
<p class="font-body-sm text-secondary">Images only (PNG, JPG, JPEG) up to 10MB</p>
</div>
</div>
</div>
<div class="flex flex-col gap-3">
<div class="flex items-center justify-between">
<span class="font-label-bold text-on-surface">TRA Certificate</span>
<span class="font-label-caps text-on-secondary-container bg-secondary-container px-2 py-0.5 rounded">MANDATORY</span>
</div>
<div class="border-2 border-dashed border-outline-variant rounded-lg p-10 bg-surface flex flex-row items-center gap-6">
<div class="w-24 h-24 bg-surface-container-highest rounded border border-outline-variant flex items-center justify-center overflow-hidden">
<img alt="Document thumbnail" class="object-cover w-full h-full opacity-60" data-alt="A macro close-up of a corporate certificate with a faint holographic seal and crisp black typography on premium textured cream paper. The lighting is soft and professional, highlighting the high-stakes legal nature of the document. The scene is shot with a shallow depth of field, emphasizing the official branding and security features in a modern light-mode office setting." src="https://lh3.googleusercontent.com/aida-public/AB6AXuC6_90nv32pdYd5Pj2GoLlDIsbmPi4hiqiWbta57hHhOayMvH66H3HerbrzvYZfxFVwEuOxESCt2_TGCgFztwyZclHl7jSKFrKp2Oa9RUhOJUjvKQtMFoF0XdA1BvGiKPiuFpiadMeyfm451Ae6oS-Y0Q1FUwDquh0r2OoysMFwRoIqub-le-imbMvX7oRopAtCRHjumpi5wx0S32Y9H6mGZRxtlnF6QWbjLoFWe8AZlUk-qlkdKk7epRFe9aR37yWZgcKKOVRV1xY"/>
</div>
<div class="flex-1">
<div class="flex items-center justify-between mb-1">
<p class="font-label-bold text-on-surface">TRA_Cert_2024.jpg</p>
<button class="text-error hover:bg-error-container p-1 rounded-full transition-colors">
<span class="material-symbols-outlined text-[20px]" data-icon="delete">delete</span>
</button>
</div>
<div class="flex items-center gap-2 mb-2">
<span class="material-symbols-outlined text-on-tertiary-container text-[18px]" data-icon="check_circle" data-weight="fill" style="font-variation-settings: 'FILL' 1;">check_circle</span>
<span class="font-body-sm text-on-tertiary-container">Upload Complete</span>
</div>
<div class="w-full bg-surface-container-highest h-1 rounded-full">
<div class="bg-on-tertiary-container h-full w-full rounded-full"></div>
</div>
</div>
</div>
</div>
</div>
</div>
<div class="flex items-center justify-between pt-4">
<button class="flex items-center gap-2 px-6 h-12 font-label-bold text-on-surface border border-outline-variant rounded hover:bg-surface-container transition-all">
<span class="material-symbols-outlined text-[20px]" data-icon="arrow_back">arrow_back</span>
                        Back
                    </button>
<button class="flex items-center gap-2 px-8 h-12 font-label-bold text-white bg-on-tertiary-container rounded hover:opacity-90 transition-all shadow-md">
                        Complete Registration
                        <span class="material-symbols-outlined text-[20px]" data-icon="done_all">done_all</span>
</button>
</div>
</section>
<aside class="lg:col-span-5 space-y-6">
<div class="bg-primary-container p-8 rounded-lg text-white">
<h3 class="font-label-caps tracking-widest text-on-primary-container mb-4">GUIDANCE</h3>
<div class="space-y-6">
<div class="flex gap-4">
<span class="material-symbols-outlined text-on-tertiary-fixed-dim" data-icon="info">info</span>
<div class="flex-1">
<p class="font-label-bold mb-1">BRELA Requirements</p>
<p class="text-on-primary-container text-sm leading-relaxed">Ensure the Business Registration and Licensing Authority (BRELA) certificate shows a clear stamp and the current date of incorporation.</p>
</div>
</div>
<div class="flex gap-4">
<span class="material-symbols-outlined text-on-tertiary-fixed-dim" data-icon="security">security</span>
<div class="flex-1">
<p class="font-label-bold mb-1">Data Privacy</p>
<p class="text-on-primary-container text-sm leading-relaxed">Your documents are encrypted using AES-256 standards before being stored in our secure vault.</p>
</div>
</div>
</div>
</div>
<div class="bg-white border border-outline-variant rounded-lg overflow-hidden">
<div class="relative h-48">
<img alt="Support Office" class="object-cover w-full h-full" data-alt="A wide-angle, bright interior shot of a high-end, minimalist corporate headquarters lobby with glass walls and clean white surfaces. The atmosphere is quiet, professional, and institutional, conveying a sense of stability and institutional reliability. High-key natural lighting floods the space, highlighting the pristine architectural lines and the absence of clutter, fitting the modern light-mode brand identity." src="https://lh3.googleusercontent.com/aida-public/AB6AXuAMqTWRPq2UBDAIhINOI538euqzilYK7cm08vWyFj2xF_yn2CBSZnLBnIxdALrlj69JIbxia56aoMS2oF01CgWf45cT7g4oP7bfk2-j3gY2Cg65s8-ytevpfv6aPLSErQtNMWBTepcRTudtEJBnYMPe4md3JnDIQpu-GMc_xAZMe7O_n-i0rc1hcw2MVCdyhf7tVSZqsfm91ion4yYnGh1D-dVtk5AGE9iD_evcr_-fczsyIqbj0lY5aFj3uoF4Mn0rvc-R9tfNXJs"/>
<div class="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent flex flex-col justify-end p-6">
<p class="text-white font-label-bold">Need help with registration?</p>
<p class="text-white/80 text-sm">Our compliance officers are available 24/7</p>
</div>
</div>
<div class="p-6">
<button class="w-full h-10 border border-outline rounded font-label-bold text-on-surface hover:bg-surface transition-all flex items-center justify-center gap-2">
<span class="material-symbols-outlined text-[18px]" data-icon="headset_mic">headset_mic</span>
                            Speak to an Advisor
                        </button>
</div>
</div>
</aside>
</div>
</main>
<footer class="mt-20 py-12 bg-surface-container-low border-t border-outline-variant">
<div class="max-w-[1120px] mx-auto px-6 flex flex-col md:flex-row justify-between items-center gap-8">
<div class="flex flex-col gap-2">
<p class="font-label-bold text-on-surface">TrustCore Portal</p>
<p class="text-secondary text-sm">© 2024 Institutional Verification Services. All rights reserved.</p>
</div>
<div class="flex gap-8">
<a class="text-secondary hover:text-on-surface text-sm transition-colors" href="#">Privacy Policy</a>
<a class="text-secondary hover:text-on-surface text-sm transition-colors" href="#">Terms of Service</a>
<a class="text-secondary hover:text-on-surface text-sm transition-colors" href="#">Compliance Standards</a>
</div>
</div>
</footer>
</body></html>