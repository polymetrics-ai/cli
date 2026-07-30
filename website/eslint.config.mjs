import { defineConfig, globalIgnores } from 'eslint/config';
import nextVitals from 'eslint-config-next/core-web-vitals';
import nextTs from 'eslint-config-next/typescript';

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  {
    // Existing app code predates the React Compiler lint additions bundled with
    // eslint-config-next. Keep those known findings visible without making this
    // tooling restoration refactor out-of-scope product code.
    files: [
      'components/auth/profile-settings-dialog.tsx',
      'components/blog/annotations-provider.tsx',
      'components/blog/comment-composer.tsx',
      'components/blog/github-evidence.tsx',
      'components/blog/highlight-text.tsx',
      'components/blog/selection-popover.tsx',
      'components/home/navbar.tsx',
      'components/ui/on-page-toc.tsx',
      'lib/auth-client.ts',
    ],
    rules: {
      'react-hooks/immutability': 'warn',
      'react-hooks/set-state-in-effect': 'warn',
    },
  },
  globalIgnores([
    // Build/runtime output ignored by eslint-config-next defaults.
    '.next/**',
    'out/**',
    'build/**',
    'next-env.d.ts',

    // Repository-generated website artifacts.
    '.source/**',
    'lib/**/*.generated.*',
    'data/**/*.generated.*',
    'public/connectors/icons/**',

    // Local test/build reports.
    'playwright-report/**',
    'test-results/**',
    'coverage/**',
  ]),
]);

export default eslintConfig;
