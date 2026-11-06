import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

test.describe('Reportes de asistencia', () => {
  test('muestra el reporte diario, el panel y descarga exportaciones', async ({ page }) => {
    await ingresarComoAdministrador(page);
    await page.goto('/reportes/diario');

    await expect(page.getByRole('heading', { name: 'Reporte diario' })).toBeVisible();
    await expect(page.getByText('Presentes').first()).toBeVisible();
    await expect(page.getByRole('button', { name: 'Filtrar' })).toBeVisible();

    await page.getByLabel('Estado').selectOption({ label: 'Tardanza' });
    await page.getByRole('button', { name: 'Filtrar' }).click();
    await expect(page).toHaveURL(/estado=tardanza/);

    const [excel] = await Promise.all([
      page.waitForEvent('download'),
      page.getByRole('link', { name: 'Excel' }).click(),
    ]);
    expect(excel.suggestedFilename()).toContain('.xlsx');

    const [pdf] = await Promise.all([
      page.waitForEvent('download'),
      page.getByRole('link', { name: 'PDF' }).click(),
    ]);
    expect(pdf.suggestedFilename()).toContain('.pdf');
  });

  test('muestra el panel con actualización y el reporte de reiteración', async ({ page }) => {
    await ingresarComoAdministrador(page);
    await page.goto('/reportes/panel');

    await expect(page.getByRole('heading', { name: 'Panel del día' })).toBeVisible();
    await expect(page.getByText('Pendientes').first()).toBeVisible();

    await page.goto('/reportes/resumen');
    await expect(page.getByRole('heading', { name: 'Reporte acumulado' })).toBeVisible();

    await page.goto('/reportes/reiteracion');
    await expect(page.getByRole('heading', { name: 'Alumnos con reiteración' })).toBeVisible();
  });
});
