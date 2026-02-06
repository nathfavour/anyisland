import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */
const sidebars: SidebarsConfig = {
  // By default, Docusaurus generates a sidebar from the docs folder structure
  tutorialSidebar: [
    {type: 'doc', id: 'intro', label: 'Introduction'},
    {type: 'doc', id: 'architecture', label: 'Architecture'},
    {type: 'doc', id: 'manifest-spec', label: 'Manifest Specification'},
    {
      type: 'category',
      label: 'System Lifecycle',
      items: [
        'lifecycle/index',
        'lifecycle/distribution',
        'lifecycle/installation',
        'lifecycle/updates',
        'lifecycle/audit',
        'lifecycle/removal',
      ],
    },
    {type: 'doc', id: 'distribution', label: 'Distribution & Integration'},
    {type: 'doc', id: 'security', label: 'Security & Privacy'},
    {type: 'doc', id: 'cli', label: 'CLI Reference'},
  ],
};

export default sidebars;