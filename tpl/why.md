{{template "%%top.gohtml" .}}

Why I made GoatCounter
======================
Last year I was working on a product idea and wanted to add some basic analytics
to measure how many people are visiting the site. I've also been wanting to add
basic analytics to my personal homepage/programming weblog to measure if anyone
is reading anything I write (and if so, what?)

Analytics are useful to measure things like *“what type of content is popular,
and should I write more of?”*, *“does it even make sense to distribute a
newsletter?”*, *“how does the redesigned signup button affect signup rates?”*,
*“is anyone even using this page I'm maintaining”?*

I tried a number of existing solutions, and found them are either very complex
and designed for advanced users, or far too simplistic. In addition almost all
hosted solutions are priced for business users (≥$10/month), making it too
expensive for personal/hobby use.

What seems to be lacking is a “middle ground” that offers useful statistics to
answer business questions, without becoming a specialized marketing tool
requiring in-depth training to use effectively. Furthermore, some tools have
privacy issues (especially Google Analytics). I saw there was space for a new
service and ended up putting my original idea in the freezer and writing
GoatCounter.[^gc]

[^gc]: A little context on the name: GoatCounter is written in the Go
       programming language, and I thought it would be fun to reflect that in
       the name. The original “intermediate” project in-between my original idea
       and GoatCounter was GoatLetter, a newsletter service with similar
       aesthetics to GoatCounter (something I will finish *soon™*). Probably
       subconsciously influenced by MailChimp I ended up with “Goat”.<br><br>
       I originally wanted to avoid using the word “Analytics” as it's 1)
       associated with invasive tracking like GA 2) something I have trouble
       spelling correctly 😅 “Counter” refers to “counting requests” (as opposed
       to “analytics”. It's a bit of a weird name, but memorable, so I guess
       I'll stick with it for now :-)


Why is it free?
---------------
Almost all hosted solutions are exclusively oriented towards business use. This
makes sense from a business point of view – better to support 100 customers
paying €30 each than 1000 paying €3 each – but it does leave a lot of people
without a good/affordable solution.

I think it’s important to make the barrier of entry for software like this low
as feasible to make actual meaningful inroads to “de-Google-fi” the internet a
bit, and make pervasive tracking less common. Making it freely available (for
personal use) is part of that. In my own online purchasing behaviour I find that
even a small €1 or €2 subscription is quite a barrier, especially for personal
projects. From what I see, I don’t think my behaviour is an outlier. Most people
don't use Google Analytics because they're overwhelmingly impressed by it, but
just because it's free.

The only other options outside of Google Analytics is to pay upwards of
€10/month or to self-host something like Matomo, which also isn't free in terms
of hosting costs, setup time, maintenance, etc. Never mind that average person
running his photography website probably doesn't have the interest or know-how.

If you want to make the internet a bit better, then the only real option is to
offer a SaaS for free, at least for personal use. Ideally I'd like to make it
free for *everyone* up to *n* pageviews/month – like Google Analytics – but I do
need to pay the bills 😅

What are GoatCounter's goals?
-----------------------------
Without focusing too much on specific features, high-level goals are:

- Give useful data while respecting people's privacy. For the most part, it
  should just “count events” rather than “get as much data as technically
  possible” (which, for the most part, is not even that useful or valuable for
  analytics anyway).

- There should always be an option to add GoatCounter to your site *without*
  requiring a GDPR consent notice.

- Easy user interface; some existing solutions are surprisingly complex, to the
  point where I wasn't able to get some basic data out of it. It's like putting
  a layperson in front of a SQL database and telling them to “just” get some
  “simple” data out of it.

  GoatCounter isn't intended to solve every possible analytics use case, and by
  limiting the scope it should be *better* for the use cases it *is* designed
  for.

- Make a web app that I *like* using, rather than merely *tolerate*. This is a
  bit subjective, and perhaps my tastes are old-fashioned, but I'm not wildly
  impressed by a lot of modern web UIs. Google Analytics is a good example where
  pressing the “back button” will often break everything.

- Works well with any browser and assistive technology, whenever reasonably
  possible.

- Easy to self-host without too much mucking about with web servers, proxies,
  {PHP,Python,Ruby,NodeJS,...}, SQL databases, what-have-you. I feel this is an
  important feature, because “run your own” sounds nice but it becomes a bit of
  a niche feature if you need to have a lot of knowledge and spend a lot of time
  setting everything up.

---

**Footnotes**

* Footnotes
{:footnotes}

{{template "%%bottom.gohtml" .}}
