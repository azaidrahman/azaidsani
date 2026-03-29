## Why arent there better tools for financing?
Something I've always been meaning to do, is to be able to not rely on a third
party app to manage my finances. I have been using [Money
Manager](https://www.realbyteapps.com) for the longest time. Dont get me wrong,
its a great app. But I feel so limited in that, if I wanted to extract some kind
of insight that wasnt already available in the app, I would have to go through
so many hoops and exports/etc. 

In the end I end up not doing anything. I just record my daily expenses blindly.
Next year, I will have to start paying for a house, and I just realized if I
dont have a keen eye on my finances now, the hammer will fall and it wont fall
gently.  

Introducing, **tele-expenses**! 

So I built a telegram bot that not only track my expenses, but it gives me
visibility on all my debts and spending habits and actively adjusts the budget
according to my needs.

## Source
I was inspired by this [post](https://x.com/mrkaran_/status/2035360975080370216) which spoke to me on many levels.

1. I love the file > app philosophy.
2. I wanted something accounting-like, and robust. Not a spreadsheet. Not a
   database.

So decided to build something that was simple, and inspired by my last bot on
slack, I decided to try it on telegram, which i already use. 

## What it can do
The app works by simply writing to a hledger file in a google cloud storage on
GCS. Then I have a couple of yaml files that store the data on the categories,
finances, etc.  

So the actual usage of it is simple typing `/add` on the bot and it pops up as a
guided wizard. Its abit of work and it lacks autocomplete. But im happy it does
the job.

## Features
The best part is I can do any accounting tricks I like. I can set up regular
credit and debit accounts among friends so we can debts more reasonably in an
accounting manner. This would be needlessly difficult on a spreadsheet (i've
tried). 

I can also introduce rolling budgets, so if on any given category that doesnt
have any spending, it can roll over to the next month for 3 months. So i can
keep lets say a budget on apparel relatively low, but overtime it grows.

TBH I dont know if other apps have this but man have I always wanted it.

Finally, I have a easy way to introduce installments. I just set up the hledger
by crediting from the Shoppee Account. Then setup multiple future transactions
that eventually 0's out that debt.

Its so fun to have robust accounting tools, all as a simple file.

Cant wait to try out more.

