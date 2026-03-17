---
title: "How I Built This Website"
date: 2026-03-16
draft: false
tags: ["hugo", "cloudflare", "ai"]
---

I've been meaning to put a personal site together for a while. Not anything fancy, just somewhere to write and put my thoughts out. Between work and everything else, it kept getting pushed back. So I decided to stop overthinking it and just get something live.

## The stack

The site runs on [Hugo](https://gohugo.io/), a static site generator written in Go. I went with Hugo because it's fast, simple, and I'm learning Go anyway so I figured I'd stay in that ecosystem. The theme is [risotto](https://github.com/joeroe/risotto) — clean, text-forward, no distractions. Exactly what I wanted.

For hosting I'm using Cloudflare Pages. The CI/CD is dead simple: push to `main` and Cloudflare builds and deploys automatically. No pipelines to maintain, no servers to babysit. For someone who does DevOps during the day, it's nice to not have to think about infrastructure for once.

## AI did most of the heavy lifting

To be honest, I used AI (Claude) to build most of this. The Hugo config, the project structure, getting the theme wired up, the deployment setup. I described what I wanted and worked through it conversationally. I'm not a frontend developer and I didn't want to spend hours reading Hugo docs for what's ultimately a simple static site.

That said, I'm not trying to hand everything off to AI and call it a day. I'm using this site as a way to actually learn web development and CSS on my own terms. The AI helped me get the scaffolding up fast so I could focus on understanding the pieces rather than fighting with boilerplate.

## What's next

Eventually I want to move away from Hugo and host this myself, probably a Go server that serves the content directly. I want to understand the full stack, from the server to the HTML. But that's a project for when I have more time. For now, this gets the job done: I have somewhere to write, it deploys itself, and I can iterate on it whenever I get a chance.

The whole point was to stop waiting for the perfect setup and just start writing. So here we are.
