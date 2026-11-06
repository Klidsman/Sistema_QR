import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

function hoyISO(): string {
  return new Date().toISOString().slice(0, 10);
}

test.describe('Historia 7: cursos y horario semanal', () => {
  test('crea un curso y lo asigna al horario de una sección', async ({ page }) => {
    await ingresarComoAdministrador(page);
    const sufijo = Date.now().toString().slice(-4);
    const curso = `Matemática ${sufijo}`;

    await page.goto('/configuracion/turnos');
    const horario = page.locator('form[action="/configuracion/turnos/manana"]');
    await horario.getByLabel('Hora de ingreso (Mañana) *').fill('07:30');
    await horario.getByLabel('Tolerancia en minutos (Mañana) *').fill('10');
    await horario.getByLabel('Hora límite de tardanza (Mañana) *').fill('08:00');
    await horario.getByLabel('Hora de cierre (Mañana) *').fill('08:30');
    await horario.getByLabel('Vigencia desde (Mañana) *').fill(hoyISO());
    await horario.getByRole('button', { name: 'Guardar horario' }).click();
    await expect(page.getByText('El horario del turno se guardó correctamente.')).toBeVisible();

    await page.goto('/configuracion/bloques?turno=manana');
    await page.getByLabel('Número de bloque *').fill('1');
    await page.getByLabel('Hora de inicio *').fill('07:45');
    await page.getByLabel('Hora de fin *').fill('08:30');
    await page.getByLabel('Vigencia desde *').fill(hoyISO());
    await page.getByRole('button', { name: 'Agregar bloque' }).click();
    await expect(page.getByText('El bloque de clase se creó correctamente.')).toBeVisible();

    await page.goto('/configuracion/cursos');
    await page.getByLabel('Nombre del curso *').fill(curso);
    await page.getByLabel('1.°', { exact: true }).check();
    await page.getByRole('button', { name: 'Agregar curso' }).click();
    await expect(page.getByText('El curso se creó correctamente.')).toBeVisible();
    await expect(page.getByText(curso)).toBeVisible();

    await page.goto('/configuracion/horarios');
    await page.getByLabel('Sección').selectOption({ label: '1.° A' });
    await page.getByLabel('Fecha').fill(hoyISO());
    await page.getByRole('button', { name: 'Ver horario' }).click();

    await expect(page.getByRole('heading', { name: /1\.° A/ })).toBeVisible();
    const celda = page.getByLabel('Lunes bloque 1');
    await celda.selectOption({ label: curso });
    await page.getByRole('button', { name: 'Guardar horario' }).click();
    await page.waitForLoadState('load');

    const confirmar = page.getByRole('button', { name: 'Confirmar reemplazo' });
    if ((await confirmar.count()) > 0) {
      await confirmar.click();
    }

    await expect(page.getByText('El horario semanal se guardó correctamente.')).toBeVisible();
    const celdaGuardada = page.getByLabel('Lunes bloque 1');
    await expect(celdaGuardada.locator('option:checked')).toHaveText(curso);
  });
});
