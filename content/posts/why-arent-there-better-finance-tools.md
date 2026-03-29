---
title: Why arent there better finance tools?
date: 2026-03-29T00:00:00Z
draft: true
tags:
    - ai
    - finance
    - telegram
---
{{< movies src="/images/ntbtstm-feature.jpg" caption="Nirvanna The Band The Show The Movie" >}}
## Why arent there better tools for personal finance?
Ive always wanted to manage my finances without relying on a third party app.
Ive been using [Money Manager](https://www.realbyteapps.com) for years now.
Dont get me wrong, its a great app. But the moment I want any insight that isnt
already baked in, I have to jump through hoops with exports and workarounds.

So I end up not doing anything. I just record my daily expenses blindly. Next
year I start paying for a house, and if I dont have a keen eye on my finances
now, the hammer will fall and it wont fall gently.

Introducing, **tele-expenses**!

So I built a Telegram bot that tracks my expenses, gives me visibility on all my
debts and spending habits, and actively adjusts the budget according to my needs.

## Source
I was inspired by this [post](https://x.com/mrkaran_/status/2035360975080370216) which spoke to me on many levels.

1. I love the file > app philosophy.
2. I wanted something accounting-like and robust. Not a spreadsheet. Not a
   database.

So I decided to build something simple. Inspired by a previous bot I built on
Slack, I went with Telegram since I already use it.

## Why hledger
The backbone of this whole thing is [hledger](https://hledger.org). Its a
plaintext accounting tool that uses *double-entry bookkeeping*. All your
transactions live in a single journal file that you can read and edit with any
text editor. Super simple.

Every transaction has balanced debits and credits, just like real accounting. And
because its all plaintext, I can version control it, script against it, and
generate reports from it however I want. Its the kind of tool that scales from
simple expense tracking all the way to proper multi-account bookkeeping without
ever getting in the way.

## What it can do
The app works by writing to an hledger file stored on Google Cloud Storage. I
also have a couple of YAML files for categories, budgets, and other config.

The actual usage is simple. Type `/add` on the bot and it walks you through a
guided wizard. Its a bit of work and it lacks autocomplete, but it gets the job
done.

## Features
{{< mid-img src="/images/screenshot-teleexpense-20260329-1101pm.png" caption="tele-expense /help screen" >}}
The best part is I can do any accounting tricks I like. I can set up credit and
debit accounts between friends so we can settle debts properly using actual
accounting. This would be needlessly difficult on a spreadsheet (Ive tried).

I can also set up rolling budgets. If a category has no spending in a given
month, the budget rolls over for up to 3 months. So I can keep something like an
apparel budget relatively low, but over time it grows.

TBH I dont know if other apps have this, but Ive always wanted it.

Finally, I have an easy way to handle installments. I set up hledger to credit
from the Shopee account, then create future transactions that eventually zero
out the debt.

Its so fun to have robust accounting tools, all as a simple file.

Cant wait to do more with it.

