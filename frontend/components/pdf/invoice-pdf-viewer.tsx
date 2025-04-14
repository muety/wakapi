import { Document, Page, StyleSheet, Text, View } from "@react-pdf/renderer";
import { format } from "date-fns";
import React from "react";

import { Invoice } from "@/lib/types";
import { formatCurrency, formatNumber, getHours } from "@/lib/utils";

interface iProps {
  invoiceData: Invoice;
}

export const InvoicePDFViewer = ({ invoiceData }: iProps) => {
  const {
    client,
    line_items,
    tax,
    final_message,
    heading,
    origin,
    destination,
    created_at: date,
  } = invoiceData;

  const totalInvoice = React.useMemo(() => {
    return line_items.reduce((acc, item) => {
      return acc + getHours(item.total_seconds) * client.hourly_rate;
    }, 0);
  }, [line_items, client.hourly_rate]);

  const taxTotal = React.useMemo(() => {
    if (isNaN(tax)) {
      return 0;
    }
    return totalInvoice * (tax / 100);
  }, [totalInvoice, tax]);

  const netTotal = React.useMemo(() => {
    return totalInvoice + taxTotal;
  }, [totalInvoice, taxTotal]);

  const Header = () => (
    <View style={styles.header}>
      <View>
        <Text style={styles.invoiceTitle}>INVOICE</Text>
        <Text style={{ fontSize: 11, ...styles.mutedText }}>
          {invoiceData.invoice_summary}
        </Text>
      </View>
      <View style={styles.flexRow}>
        <View>
          <Text style={{ fontSize: 12, fontWeight: 700 }}>Invoice #:</Text>
          <Text style={{ fontSize: 12, fontWeight: 700 }}>Date:</Text>
        </View>
        <View style={styles.flexColEnd}>
          <Text style={{ fontSize: 11, ...styles.mutedText }}>
            INV-{invoiceData.invoice_id}
          </Text>
          <Text style={{ fontSize: 11, ...styles.mutedText }}>
            {format(date, "MMM dd, yyyy")}
          </Text>
        </View>
      </View>
    </View>
  );

  const TitleContent = ({
    title,
    content,
  }: {
    title: string;
    content: string;
  }) => (
    <View
      style={{
        display: "flex",
        alignItems: "flex-start",
        gap: "4px",
        marginBottom: 12,
      }}
    >
      <Text style={{ fontSize: 10, fontWeight: 700 }}>{title}</Text>
      <Text style={{ fontSize: 10, ...styles.mutedText }}>{content}</Text>
    </View>
  );

  return (
    <Document title="Invoice">
      <Page style={styles.page} size="A4">
        <View style={styles.main}>
          <Header />
          <TitleContent title="From:" content={origin} />
          <TitleContent title="To:" content={destination} />
          {heading && (
            <View
              style={{
                display: "flex",
                alignItems: "flex-start",
                gap: "4px",
                marginTop: "8px",
                marginBottom: "4px",
              }}
            >
              <Text style={{ fontSize: 10, fontWeight: 700 }}>{heading}</Text>
            </View>
          )}
          <View style={styles.table}>
            <View style={[styles.tableRow, styles.tableHeader]}>
              <View style={styles.tableCol}>
                <Text style={styles.tableCell}>Project/Item</Text>
              </View>
              <View style={styles.tableCol}>
                <Text style={styles.tableCell}>
                  Hourly Rate ({client.currency})
                </Text>
              </View>
              <View style={styles.tableCol}>
                <Text style={styles.tableCell}>Qty (Hrs)</Text>
              </View>
              <View style={styles.tableCol}>
                <Text style={styles.tableCell}>Amount ({client.currency})</Text>
              </View>
            </View>
            {line_items?.map((item) => (
              <View style={styles.tableRow} key={item.title}>
                <View style={styles.tableCol}>
                  <Text style={styles.tableCell}>{item.title}</Text>
                </View>
                <View style={styles.tableCol}>
                  <Text style={styles.tableCell}>
                    {formatCurrency(client.hourly_rate, client.currency)}
                  </Text>
                </View>
                <View style={styles.tableCol}>
                  <Text style={styles.tableCell}>
                    {formatNumber(getHours(item.total_seconds))}
                  </Text>
                </View>
                <View style={styles.tableCol}>
                  <Text style={styles.tableCell}>
                    {formatCurrency(
                      client.hourly_rate * getHours(item.total_seconds),
                      client.currency
                    )}
                  </Text>
                </View>
              </View>
            ))}
          </View>
          <View
            style={{
              display: "flex",
              flexDirection: "row",
              justifyContent: "flex-end",
              marginTop: 12,
              gap: 12,
            }}
          >
            <View style={styles.flexColStart}>
              <Text style={styles.invoiceTotalsTitle}>Subtotal</Text>
              <Text style={styles.invoiceTotalsTitle}>Tax</Text>
              <Text style={styles.invoiceTotalsTitle}>Total</Text>
            </View>
            <View style={{ ...styles.flexColEnd }}>
              <Text style={styles.invoiceTotalsValue}>
                {formatCurrency(totalInvoice, client.currency)}
              </Text>
              <Text style={styles.invoiceTotalsValue}>
                {formatCurrency(taxTotal, client.currency)}
              </Text>
              <Text style={styles.invoiceTotalsValue}>
                {formatCurrency(netTotal, client.currency)}
              </Text>
            </View>
          </View>
          <View
            style={{
              display: "flex",
              alignItems: "flex-start",
              gap: "4px",
            }}
          >
            <Text style={{ fontSize: 10, fontWeight: 700 }}>
              {final_message}
            </Text>
          </View>
        </View>
      </Page>
    </Document>
  );
};

const styles = StyleSheet.create({
  viewer: {
    paddingTop: 32,
    width: "100%",
    height: "80vh",
    border: "none",
  },
  page: {
    display: "flex",
    padding: "0.4in 0.4in",
    fontSize: 10,
    color: "#333",
    backgroundColor: "#fff",
  },
  tableHeader: {
    backgroundColor: "rgb(250, 250, 250)",
    color: "#000000",
    fontSize: 16,
    fontWeight: 700,
    // fontFamily: "'__Rubik_b539cb', '__Rubik_Fallback_b539cb'",
  },
  table: {
    display: "flex",
    width: "auto",
    // borderStyle: "solid",
    borderWidth: 1,
    borderRightWidth: 0,
    borderBottomWidth: 0,
    marginTop: 24,
    // borderColor: "#d9d9d9",
    border: "1px solid rgb(217, 217, 217)",
    borderRadius: "8px",
  },
  tableRow: {
    margin: "auto",
    flexDirection: "row",
  },
  mutedText: {
    color: "gray",
  },
  tableCol: {
    width: "25%",
    borderStyle: "solid",
    borderBottomWidth: 1,
    borderRightWidth: 1,
    borderLeftWidth: 0,
    borderTopWidth: 0,
    borderColor: "#d9d9d9",
  },
  tableCell: {
    margin: "auto",
    paddingTop: 8,
    paddingBottom: 8,
    fontSize: 10,
  },
  divider: {
    width: "100%",
    height: 1,
    marginTop: 32,
    marginBottom: 32,
    backgroundColor: "#D9D9D9",
  },
  totalDivider: {
    width: "100%",
    height: 1,
    marginTop: 2,
    marginBottom: 2,
    backgroundColor: "#D9D9D9",
  },
  main: {
    transitionProperty: "box-shadow, transform",
    transitionDuration: "0.4s",
    transitionTimingFunction: "cubic-bezier(0.19, 1, 0.22, 1)",
    overflow: "hidden",
    width: "100%",
    height: "100%",
    paddingBottom: 60,
    paddingLeft: 40,
    paddingRight: 40,
    backgroundColor: "hsl(0deg 0% 100%)",
    boxShadow:
      " 0 0 0 1px rgba(0, 0, 0, 0.05), 0 7px 25px 0 rgba(0, 0, 0, 0.03), 0 4px 12px 0 rgba(0, 0, 0, 0.03)",
  },
  header: {
    display: "flex",
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 24,
  },
  invoiceTitle: {
    fontSize: 30,
    color: "#333",
  },
  flexRow: {
    display: "flex",
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
  },
  flexCol: {
    display: "flex",
    flexDirection: "column",
    justifyContent: "space-between",
    alignItems: "center",
  },
  flexColEnd: {
    display: "flex",
    flexDirection: "column",
    justifyContent: "flex-end",
    alignItems: "flex-end",
  },
  flexColStart: {
    display: "flex",
    flexDirection: "column",
    justifyContent: "flex-end",
    alignItems: "flex-start",
  },
  invoiceTotalsTitle: {
    fontSize: 11,
    fontWeight: 700,
  },
  invoiceTotalsValue: {
    fontSize: 11,
    color: "hsl(0deg 0.42% 53.53%)",
  },
});
