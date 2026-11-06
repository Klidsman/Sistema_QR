import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

function hoyISO(): string {
  return new Date().toISOString().slice(0, 10);
}

test.describe('Exportación de asistencia', () => {
  test('descarga un CSV con el encabezado esperado', async ({ page }) => {
    await ingresarComoAdministrador(page);

    await page.goto('/asistencia/exportar');
    await expect(page.getByRole('heading', { name: 'Exportar asistencia' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Descargar CSV' })).toBeVisible();

    const respuesta = await page.request.get(
      `/asistencia/exportar/archivo?desde=${hoyISO()}&hasta=${hoyISO()}&seccion_id=1`,
    );
    expect(respuesta.status()).toBe(200);
    expect(respuesta.headers()['content-type']).toContain('text/csv');
    const contenido = await respuesta.text();
    expect(contenido).toContain('identificador');
    expect(contenido).toContain('docente');
  });
});
