import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { TOP_FAQS } from "@/lib/constants/faqs";

export function FAQ() {
  return (
    <div
      className="flex flex-col justify-center align-middle m-auto px-14 mx-14 text-md"
      style={{ minHeight: "70vh" }}
    >
      <h1 className="text-6xl mb-8 text-center">FAQs</h1>
      <Accordion type="single" collapsible>
        {TOP_FAQS.map((faq) => (
          <AccordionItem value={faq.question} key={faq.question}>
            <AccordionTrigger>{faq.question}</AccordionTrigger>
            <AccordionContent>{faq.answer}</AccordionContent>
          </AccordionItem>
        ))}
      </Accordion>
    </div>
  );
}
