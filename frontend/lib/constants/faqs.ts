export interface Faq {
  question: string;
  answer: string;
}

export const TOP_FAQS: Faq[] = [
  {
    question: "The What?",
    answer:
      "This is a self-hosted version of wakatime.com that is open source and free to use. We rely on the same open source plugins and collect the same data that is available from wakatime's open source plugins.",
  },
  {
    question: "What this isn't?",
    answer:
      "It is not a way to measure the productivity of developers. Don't be a jerk.",
  },
  {
    question: "What is the difference between this and Wakapi?",
    answer:
      "Wakana is similar to wakapi but with a more commercial focus. I've ditched the horrible golang templates for some next.js but besides that this project is 90% wakapi backend. We plan to add more features and a more polished UI as well as focus on greater feature parity with wakatime",
  },
  {
    question: "Why is this open source?",
    answer:
      "This uses the wakapi codebase as a starting point. Wakapi is open source and free to use. We rely on wakatime's open source plugins and collect the same data that is available from the plugins. This project cannot simply be proprietary. Unless you're an ungrateful bastard that prefers to benefit more from the work of others without giving back.",
  },
  {
    question: "Caution: This is a work in progress",
    answer:
      "This is a work in progress. We are not yet ready for production use. We are still working on the UI and adding features. We are also working on a managed version of this website that will be available to paying customers to help fund the continued work on this project.",
  },
  {
    question: "Would I like to work at wakatime?",
    answer:
      "I won't mind working for wakatime if this doesn't work out. I just don't want to write a single line of python with server rendered templates. That thing is shit and I hated it at my first job and I'm not going to do it again. Perhaps, for a couple bucks but not for all the team in china",
  },
  {
    question: "My state of mind",
    answer:
      "It is 29th October, 2024, 00:40 GMT. I work two jobs that pay me a total of under ${430} based on my exchange rate. I'm just ranting and hope you excuse some of the language. I hope this turns out well though. And of course, one of you cool guys will clean this up eventually.",
  },
];
