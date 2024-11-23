"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import React from "react";

import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Project } from "../projects-table";
import WMultiSelect from "../w-multi-select";
import { CURRENCY_OPTIONS } from "@/lib/constants/currencies";
import { Combobox } from "../combo-box";
import { Client } from "../clients-table";
import { Icons } from "../icons";

const formSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters long"),
  currency: z.string().min(2, "Currency must be at least 2 characters long"),
  projects: z.array(z.string()).nonempty("Projects array must not be empty"),
  hourly_rate: z.preprocess((value) => {
    // Convert the value to a number if it is not already a number
    const numberValue = Number(value);
    return isNaN(numberValue) ? undefined : numberValue;
  }, z.number().min(0, "Hourly rate must be a positive number")),
});

const defaultValues = {
  name: "",
  currency: "",
  hourly_rate: 0, // Provide a sensible default value for hourly_rate
  projects: [],
};

export interface iProps {
  // type me please
  onSubmit: (values: any) => void;
  projects: Project[];
  editing: Client | null;
  loading?: boolean;
}

export function ClientForm({
  onSubmit,
  projects,
  editing,
  loading = false,
}: iProps) {
  // 1. Define your form.
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      ...defaultValues,
      ...(editing || {}),
    },
  });

  console.log("default Values", {
    ...defaultValues,
    ...(editing || {}),
  });

  // TODO: IMPLEMENT
  const onSubmitHandler = (values: z.infer<typeof formSchema>) => {
    console.log(values);
    onSubmit(values);
  };

  const PROJECT_OPTIONS = React.useMemo(() => {
    return projects.map((project) => {
      return { label: project.name, value: project.id || "" };
    });
  }, [projects]);

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmitHandler)} className="space-y-3">
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Name</FormLabel>
              <FormControl>
                <Input placeholder="Boggle" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="hourly_rate"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Hourly Rate</FormLabel>
              <FormControl>
                <Input placeholder="10.5" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="currency"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Currency</FormLabel>
              <FormControl>
                <Combobox
                  options={CURRENCY_OPTIONS}
                  defaultValue={field.value}
                  onChange={(currency: any) => {
                    form.setValue("currency", currency);
                  }}
                  placeholder="Choose Currency"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="projects"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Projects</FormLabel>
              <FormControl>
                <WMultiSelect
                  options={PROJECT_OPTIONS}
                  onSelectedOptionsChanged={(options: any[]) =>
                    form.setValue("projects", options as any)
                  }
                  defaultValue={field.value}
                  placeholder="Select projects"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <div className="flex justify-end">
          {loading && <Icons.spinner className="mr-2 size-4 animate-spin" />}
          <Button type="submit">Submit</Button>
        </div>
      </form>
    </Form>
  );
}
