import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

test.describe('Historia 3: panel de alertas', () => {
  test('muestra el panel, el indicador y aplica filtros', async ({ page }) => {
    await ingresarComoAdministrador(page);
    await page.goto('/alertas');

    await expect(page.getByRole('heading', { name: 'Alertas', exact: true })).toBeVisible();
    await expect(page.getByText('Alertas nuevas')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Filtrar' })).toBeVisible();

    await page.getByLabel('Tipo').selectOption({ label: 'Faltas seguidas' });
    await page.getByRole('button', { name: 'Filtrar' }).click();

    await expect(page).toHaveURL(/tipo=reiteracion_faltas/);
    await expect(page.getByRole('heading', { name: 'Alertas', exact: true })).toBeVisible();
  });
});
