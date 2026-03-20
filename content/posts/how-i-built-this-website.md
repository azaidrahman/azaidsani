---
title: "How I Built This Website"
date: 2026-03-16
draft: false
tags: ["hugo", "cloudflare", "ai"]
---
{{< movies src="/images/tokyo-fist.jpg" caption="Tokyo Fist (1995)" >}}
I've been meaning to put a personal site together for a while. Not anything fancy, just somewhere to write and put my thoughts out. Between work and everything else, it kept getting pushed back. So I decided to stop overthinking it and just get something live.

## The stack

The site runs on [Hugo](https://gohugo.io/), a static site generator written in Go. Something you've probably heard of before. I went with it because it's fast, simple, and I'm learning Go anyway so I figured I'd stay in that ecosystem. 

The *theme* is [risotto](https://github.com/joeroe/risotto) simply because its clean, text-forward, no distractions. Exactly what I wanted.

For *hosting* I'm using Cloudflare Pages. The CI/CD is just: push to `main` and Cloudflare builds. Why? Mostly because theres no servers to babysit. For someone who does that for a living during the day, it's nice to just keep it simple stupid.

## AI did most of the heavy lifting

To be honest, I used AI (Claude) to build most of this. The Hugo config, the project structure, getting the theme wired up, the deployment setup. I described what I wanted and worked through it conversationally. I'm not a frontend developer and I didn't want to spend hours reading Hugo docs for what's ultimately a simple static site.

That said, I'm not trying to hand everything off to AI and call it a day. I'm using this site as a way to actually learn web development and CSS on my own terms. The AI helped me get the scaffolding up fast so I could focus on understanding the pieces rather than fighting with boilerplate.

More importantly, I want to just keep writing more. Writing in public is a
pretty cool way to learn.

## What's next

Eventually I want to move away from Hugo and host this myself, probably a Go server that serves the content directly. I want to understand the full stack, from the server to the HTML. But that's a project for when I have more time. For now, this gets the job done: I have somewhere to write, it deploys itself, and I can iterate on it whenever I get a chance.

The whole point was to stop waiting for the perfect setup and just start writing. So here we are.
