---
title: "Building a Slack Bot was more than just a day of work"
date: 2026-03-20
draft: false
tags: ["slack", "jira", "ai", "devops"]
---

{{< movies src="/images/the_presidents_cake-roof_scene.webp" caption="The Presidents Cake (2025)" >}}

Tickets are a pain and a big deterrent to actual work. Someones gotta do it though. At first it was manageable, a request here, a thread there. As it went on, the less confident I was that I could keep up. 

Reminders from weeks back of a simple permission request started showing up and I realized I had to sort this out somehow.

So I decided to spend a day or two using AI to build something that would actually help.

## Start with the people, not the tool

Before writing any code, I thought about what people at work *actually use* and what they avoid. Nobody likes using Jira. But Jira is undeniably useful. It's where tickets live, where work gets tracked, where things become visible. The problem isn't Jira itself. The problem is asking people to do more things inside Jira.

So instead of adding more loops inside Jira, I thought about tapping into something people already have open all day. And that's Slack. If I could make Slack the interface and let Jira handle the backend, I'd get the structure I need without adding friction for the people making requests.

The result is a Slack bot. Users run `/gcp-request`, fill out a quick modal, and a Jira ticket gets created automatically. The summary gets posted to a shared channel, and any replies in the Slack thread sync back to Jira as comments. People stay in Slack, I get structured tickets. Everyone wins.

{{< mid-img src="/images/geronimo-1.png" caption="Jira-nimo Stilton Bot" >}}

## Influenced by Grab

I'd recently read [Grab's engineering post about agents in Slack](https://engineering.grab.com/from-firefighting-to-building) and it stuck with me. They're operating at a completely different scale, but the main idea resonated. Meet people where they already are. I figured I could try something similar, just scoped way down to my team's actual needs.

## The process mattered more than the product

I won't get into the technical details too much. It's a Python bot using Slack Bolt, deployed on Cloud Run, nothing groundbreaking. What I actually want to talk about is how I built it.

I've been using Claude Code for a while now, but for this project I decided to go deeper with obra's [superpowers](https://github.com/obra/superpowers), particularly test driven development. I'd used it before but never fully committed to the workflow. This time I did, and it changed how the whole thing came together.

Writing tests first, then letting AI generate the implementation to pass those tests. It sounds simple, but the effect is significant. Instead of reviewing AI generated code and hoping it works, you're defining the behavior upfront and verifying it automatically. The tests become the spec. The AI fills in the rest.

## The real insight: deterministic outcomes from generated AI

This led me to a bigger realization. This is a theory but, perhaps the future of AI in software engineering isn't just throwing more AI at things. It's making **deterministic outcomes** from generated AI.

I've already been doing this on a smaller scale in my day job with Terraform modules and gcloud scripts, wrapping infrastructure changes in predictable, repeatable automation. 

But superpowers brought that same discipline to application code at a speed I hadn't experienced before. Tests went from something I wrote after the fact, to something that drove the entire development loop.

I read [a post today from the OpenClaw creator](https://substack.com/@openclaw/note/c-230139405) that touched on similar ideas. The pattern is showing up everywhere. The leverage we have as humans is the scaffolding around unpredictable outcomes and ultimately decide what is a *reasonable one*.

## Where this goes

The bot is live and handling requests now. It's small, a day and a half of work. But the approach behind it is something I want to keep pushing on. AI that generates code is useful. AI that generates code inside a framework of tests and automation is something else entirely.
