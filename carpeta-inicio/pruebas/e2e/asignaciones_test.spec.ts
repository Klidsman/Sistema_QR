import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

function fechaEnDias(dias: number): string {
  const fecha = new Date();
  fecha.setDate(fecha.getDate() + dias);
  return fecha.toISOString().slice(0, 10);
}

test.describe('Asignaciones de docente', () => {
  test('asigna un titular y un reemplazo por un día', async ({ page }) => {
    await ingresarComoAdministrador(page);
    const sufijo = Date.now().toString().slice(-5);
    const curso = `Curso ${sufijo}`;
    const titular = `Titular ${sufijo}`;
    const reemplazo = `Reemplazo ${sufijo}`;

    await page.goto('/configuracion/cursos');
    await page.getByLabel('Nombre del curso *').fill(curso);
    await page.getByLabel('1.°', { exact: true }).check();
    await page.getByRole('button', { name: 'Agregar curso' }).click();
    await expect(page.getByText('El curso se creó correctamente.')).toBeVisible();

    for (const nombre of [titular, reemplazo]) {
      await page.goto('/docentes');
      await page.getByLabel('Nombre completo *').fill(nombre);
      await page.getByLabel('Correo *').fill(`${nombre.replace(' ', '').toLowerCase()}@colegio.pe`);
      await page.getByLabel('Contraseña *').fill('clave-docente-2026');
      await page.getByRole('button', { name: 'Crear docente' }).click();
      await expect(page.getByText('La cuenta de docente se creó correctamente.')).toBeVisible();
    }

    await page.goto('/asistencia/asignaciones');
    await page.getByLabel('Docente *').selectOption({ label: titular });
    await page.getByLabel('Curso *').selectOption({ label: curso });
    await page.getByLabel('Sección *').selectOption({ label: '1.° A (2026)' });
    await page.getByLabel('Desde *').fill(fechaEnDias(0));
    await page.getByRole('button', { name: 'Crear asignación' }).click();

    await expect(page.getByText('La asignación se creó correctamente.')).toBeVisible();
    const filaTitular = page.locator('tr', { hasText: titular });
    await expect(filaTitular).toBeVisible();
    await expect(filaTitular.getByRole('cell', { name: 'Titular', exact: true })).toBeVisible();

    await page.getByLabel('Docente *').selectOption({ label: reemplazo });
    await page.getByLabel('Curso *').selectOption({ label: curso });
    await page.getByLabel('Sección *').selectOption({ label: '1.° A (2026)' });
    await page.getByLabel('Desde *').fill(fechaEnDias(1));
    await page.getByLabel('Hasta (opcional)').fill(fechaEnDias(1));
    await page.getByLabel('Titular (desmarcar para reemplazo)').uncheck();
    await page.getByRole('button', { name: 'Crear asignación' }).click();

    await expect(page.getByText('La asignación se creó correctamente.')).toBeVisible();
    const filaReemplazo = page.locator('tr', { hasText: reemplazo });
    await expect(filaReemplazo).toBeVisible();
    await expect(filaReemplazo.getByRole('cell', { name: 'Reemplazo', exact: true })).toBeVisible();
  });
});
