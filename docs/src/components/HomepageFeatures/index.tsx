import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  emoji: string;
  description: ReactNode;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'AI-Powered Ingestion',
    emoji: '🤖',
    description: (
      <>
        Transform any GitHub repository into an installed tool through intelligent source analysis and automated build plan generation.
      </>
    ),
  },
  {
    title: 'Platform Agnostic',
    emoji: '🌍',
    description: (
      <>
        Built on a Platform Abstraction Layer (PAL), Anyisland provides a unified experience across Linux, macOS, and Windows.
      </>
    ),
  },
  {
    title: 'Sovereign Environment',
    emoji: '🏝️',
    description: (
      <>
        Manage your own user-space "Island" with decentralized state synchronization via Git and secure, encrypted shell history.
      </>
    ),
  },
];

function Feature({title, emoji, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <span style={{ fontSize: '5rem' }}>{emoji}</span>
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): ReactNode {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
