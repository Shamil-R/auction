<?php

use Illuminate\Support\Facades\Schema;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Database\Migrations\Migration;

class CreateHistoryTable extends Migration
{
    /**
     * Run the migrations.
     *
     * @return void
     */
    public function up()
    {
        Schema::create('history', function (Blueprint $table) {
            $table->increments('id');
            $table->string('action', 50);
            $table->text('lot');
            $table->dateTime('created_at');
            $table->unsignedInteger('lot_id');
            $table->unsignedInteger('user_id');

            $table->index('lot_id');
            $table->index('user_id');

            $table->foreign('lot_id')->references('id')->on('lots');
            $table->foreign('user_id')->references('id')->on('users');
        });
    }

    /**
     * Reverse the migrations.
     *
     * @return void
     */
    public function down()
    {
        Schema::dropIfExists('history');
    }
}
