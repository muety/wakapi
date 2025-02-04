import Image from "next/image";

import styles from "./spinner.module.css";

export const Spinner = () => {
  return (
    <div className={`${styles.fallbackSpinner} ${styles.appLoader}`}>
      <Image
        className={styles.fallbackLogo}
        src="/white-icon.svg"
        width={120}
        height={65}
        alt="logo"
      />
      <div className={styles.loading}>
        <div className={styles.effect1}></div>
        <div className={styles.effect2}></div>
        <div className={styles.effect3}></div>
      </div>
    </div>
  );
};
